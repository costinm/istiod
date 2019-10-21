package istio_agent

import (
	"fmt"
	"os"
	"strings"
	"time"

	"istio.io/istio/pkg/cmd"
	"istio.io/istio/security/pkg/nodeagent/cache"
	"istio.io/istio/security/pkg/nodeagent/sds"
	"istio.io/istio/security/pkg/nodeagent/secretfetcher"
	"istio.io/istio/security/pkg/server/monitoring"
	"istio.io/pkg/env"
	"istio.io/pkg/log"
)

const (
	// name of authentication provider.
	caProvider     = "CA_PROVIDER"
	caProviderFlag = "caProvider"

	// CA endpoint.
	caEndpoint = "CA_ADDR"

	// names of authentication provider's plugins.
	pluginNames = "PLUGINS"

	// The trust domain corresponds to the trust root of a system.
	// Refer to https://github.com/spiffe/spiffe/blob/master/standards/SPIFFE-ID.md#21-trust-domain
	trustDomain = "TRUST_DOMAIN"

	// The workload SDS mode allows node agent to provision credentials to workload proxy by sending
	// CSR to CA.
	enableWorkloadSDS = "ENABLE_WORKLOAD_SDS"

	// The ingress gateway SDS mode allows node agent to provision credentials to ingress gateway
	// proxy by watching kubernetes secrets.
	enableIngressGatewaySDS = "ENABLE_INGRESS_GATEWAY_SDS"

	// The environmental variable name for the flag which is used to indicate the token passed
	// from envoy is always valid(ex, normal 8ks JWT).
	alwaysValidTokenFlag     = "VALID_TOKEN"
	alwaysValidTokenFlagFlag = "alwaysValidTokenFlag"

	// The environmental variable name for secret TTL, node agent decides whether a secret
	// is expired if time.now - secret.createtime >= secretTTL.
	// example value format like "90m"
	secretTTL = "SECRET_TTL"

	// The environmental variable name for grace duration that secret is re-generated
	// before it's expired time.
	// example value format like "10m"
	SecretRefreshGraceDuration = "SECRET_GRACE_DURATION"

	// The environmental variable name for key rotation job running interval.
	// example value format like "20m"
	SecretRotationInterval = "SECRET_JOB_RUN_INTERVAL"

	// The environmental variable name for staled connection recycle job running interval.
	// example value format like "5m"
	staledConnectionRecycleInterval = "STALED_CONNECTION_RECYCLE_RUN_INTERVAL"

	// The environmental variable name for the initial backoff in milliseconds.
	// example value format like "10"
	InitialBackoff = "INITIAL_BACKOFF_MSEC"

	MonitoringPort  = "MONITORING_PORT"
	EnableProfiling = "ENABLE_PROFILING"
	DebugPort       = "DEBUG_PORT"
)

var (
	workloadSdsCacheOptions cache.Options
	gatewaySdsCacheOptions  cache.Options
	serverOptions           sds.Options
	gatewaySecretChan       chan struct{}
)

// Simplified SDS setup.
//
// 1. External CA: requires authenticating the trusted JWT AND validating the SAN against the JWT.
//    For example Google CA
//
// 2. Indirect, using istiod: using K8S cert.
//
// 3. Monitor mode - watching secret in same namespace ( Ingress)
//
// 4. TODO: File watching, for backward compat/migration from mounted secrets.
func StartSDS() error {
	applyEnvVars()
	gatewaySdsCacheOptions = workloadSdsCacheOptions

	if err := validateOptions(); err != nil {
		return err
	}

	stop := make(chan struct{})

	workloadSecretCache, gatewaySecretCache := newSecretCache(serverOptions)
	if workloadSecretCache != nil {
		defer workloadSecretCache.Close()
	}
	if gatewaySecretCache != nil {
		defer gatewaySecretCache.Close()
	}

	server, err := sds.NewServer(serverOptions, workloadSecretCache, gatewaySecretCache)
	if err != nil {
		log.Errorf("failed to create sds service: %v", err)
		return fmt.Errorf("failed to create sds service")
	}
	defer server.Stop()

	monitorErrCh := make(chan error)
	// Start the monitoring server.
	if monitoringPortEnv > 0 {
		monitor, mErr := monitoring.NewMonitor(monitoringPortEnv, enableProfilingEnv)
		if mErr != nil {
			return fmt.Errorf("unable to setup monitoring: %v", mErr)
		}
		go monitor.Start(monitorErrCh)
		log.Info("citadel agent monitor has started.")
		defer monitor.Close()
	}

	go exitOnMonitorServerError(monitorErrCh)

	cmd.WaitSignal(stop)

	return nil
}

// exitOnMonitorServerError shuts down Citadel agent when monitor server stops and returns an error.
func exitOnMonitorServerError(errCh <-chan error) {
	if err := <-errCh; err != nil {
		log.Errorf("Monitoring server error: %v, terminate", err)
		os.Exit(-1)
	}
}

// newSecretCache creates the cache for workload secrets and/or gateway secrets.
// Although currently not used, Citadel Agent can serve both workload and gateway secrets at the same time.
func newSecretCache(serverOptions sds.Options) (workloadSecretCache, gatewaySecretCache *cache.SecretCache) {
	if serverOptions.EnableWorkloadSDS {
		wSecretFetcher, err := secretfetcher.NewSecretFetcher(false, serverOptions.CAEndpoint,
			serverOptions.CAProviderName, true, []byte(serverOptions.VaultTLSRootCert),
			serverOptions.VaultAddress, serverOptions.VaultRole, serverOptions.VaultAuthPath,
			serverOptions.VaultSignCsrPath)
		if err != nil {
			log.Errorf("failed to create secretFetcher for workload proxy: %v", err)
			os.Exit(1)
		}
		workloadSdsCacheOptions.TrustDomain = serverOptions.TrustDomain
		workloadSdsCacheOptions.Plugins = sds.NewPlugins(serverOptions.PluginNames)
		workloadSecretCache = cache.NewSecretCache(wSecretFetcher, sds.NotifyProxy, workloadSdsCacheOptions)
	} else {
		workloadSecretCache = nil
	}

	if serverOptions.EnableIngressGatewaySDS {
		gSecretFetcher, err := secretfetcher.NewSecretFetcher(true, "", "", false, nil, "", "", "", "")
		if err != nil {
			log.Errorf("failed to create secretFetcher for gateway proxy: %v", err)
			os.Exit(1)
		}
		gatewaySecretChan = make(chan struct{})
		gSecretFetcher.Run(gatewaySecretChan)
		gatewaySecretCache = cache.NewSecretCache(gSecretFetcher, sds.NotifyProxy, gatewaySdsCacheOptions)
	} else {
		gatewaySecretCache = nil
	}
	return workloadSecretCache, gatewaySecretCache
}

var (
	pluginNamesEnv                     = env.RegisterStringVar(pluginNames, "", "").Get()
	enableWorkloadSDSEnv               = env.RegisterBoolVar(enableWorkloadSDS, true, "").Get()
	enableIngressGatewaySDSEnv         = env.RegisterBoolVar(enableIngressGatewaySDS, false, "").Get()
	alwaysValidTokenFlagEnv            = env.RegisterBoolVar(alwaysValidTokenFlag, false, "").Get()
	caProviderEnv                      = env.RegisterStringVar(caProvider, "", "").Get()
	caEndpointEnv                      = env.RegisterStringVar(caEndpoint, "", "").Get()
	trustDomainEnv                     = env.RegisterStringVar(trustDomain, "", "").Get()
	secretTTLEnv                       = env.RegisterDurationVar(secretTTL, 24*time.Hour, "").Get()
	secretRefreshGraceDurationEnv      = env.RegisterDurationVar(SecretRefreshGraceDuration, 1*time.Hour, "").Get()
	secretRotationIntervalEnv          = env.RegisterDurationVar(SecretRotationInterval, 10*time.Minute, "").Get()
	staledConnectionRecycleIntervalEnv = env.RegisterDurationVar(staledConnectionRecycleInterval, 5*time.Minute, "").Get()
	initialBackoffEnv                  = env.RegisterIntVar(InitialBackoff, 10, "").Get()
	monitoringPortEnv                  = env.RegisterIntVar(MonitoringPort, 15014,
		"The port number for monitoring Citadel agent").Get()
	debugPortEnv = env.RegisterIntVar(DebugPort, 8080,
		"Debug endpoints dump SDS configuration and connection data from this port").Get()
	enableProfilingEnv = env.RegisterBoolVar(EnableProfiling, true,
		"Enabling profiling when monitoring Citadel agent").Get()
)

func applyEnvVars() {
	serverOptions.PluginNames = strings.Split(pluginNamesEnv, ",")

	serverOptions.EnableWorkloadSDS = enableWorkloadSDSEnv

	serverOptions.EnableIngressGatewaySDS = enableIngressGatewaySDSEnv
	serverOptions.AlwaysValidTokenFlag = alwaysValidTokenFlagEnv
	serverOptions.CAProviderName = caProviderEnv
	serverOptions.CAEndpoint = caEndpointEnv
	serverOptions.TrustDomain = trustDomainEnv
	workloadSdsCacheOptions.SecretTTL = secretTTLEnv
	workloadSdsCacheOptions.SecretRefreshGraceDuration = secretRefreshGraceDurationEnv
	workloadSdsCacheOptions.RotationInterval = secretRotationIntervalEnv

	serverOptions.RecycleInterval = staledConnectionRecycleIntervalEnv

	workloadSdsCacheOptions.InitialBackoff = int64(initialBackoffEnv)

	serverOptions.DebugPort = debugPortEnv
}

func validateOptions() error {
	// The initial backoff time (in millisec) is a random number between 0 and initBackoff.
	// Default to 10, a valid range is [10, 120000].
	initBackoff := workloadSdsCacheOptions.InitialBackoff
	if initBackoff < 10 || initBackoff > 120000 {
		return fmt.Errorf("initial backoff should be within range 10 to 120000, found: %d", initBackoff)
	}

	if serverOptions.EnableIngressGatewaySDS && serverOptions.EnableWorkloadSDS &&
		serverOptions.IngressGatewayUDSPath == serverOptions.WorkloadUDSPath {
		return fmt.Errorf("UDS paths for ingress gateway and workload cannot be the same: %s", serverOptions.IngressGatewayUDSPath)
	}

	if serverOptions.EnableWorkloadSDS {
		if serverOptions.CAProviderName == "" {
			return fmt.Errorf("CA provider cannot be empty when workload SDS is enabled")
		}
		if serverOptions.CAEndpoint == "" {
			return fmt.Errorf("CA endpoint cannot be empty when workload SDS is enabled")
		}
	}
	return nil
}
