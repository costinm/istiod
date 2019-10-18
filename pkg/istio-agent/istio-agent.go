package istio_agent

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/types"

	"istio.io/istio/pilot/cmd/pilot-agent/status"
	"istio.io/istio/pilot/pkg/features"
	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pilot/pkg/proxy"
	"istio.io/istio/pilot/pkg/serviceregistry"
	"istio.io/istio/pkg/bootstrap"
	"istio.io/istio/pkg/cmd"
	"istio.io/istio/pkg/config/constants"
	"istio.io/istio/pkg/config/mesh"
	"istio.io/istio/pkg/config/validation"
	"istio.io/istio/pkg/envoy"
	"istio.io/istio/pkg/spiffe"
	"istio.io/istio/pkg/util/gogoprotomarshal"
	"istio.io/pkg/env"
	"istio.io/pkg/log"

	envoyDiscovery "istio.io/istio/pilot/pkg/proxy/envoy"

	meshconfig "istio.io/api/mesh/v1alpha1"
	networking "istio.io/api/networking/v1alpha3"
)

// WIP: refactoring the code from main pilot-agent, with added SDS.
// Currently a bit of duplication to avoid breaking pilot-agent.
var (
	tlsServerCertChain = env.RegisterStringVar(bootstrap.IstioMetaPrefix+model.NodeMetadataTLSServerCertChain, constants.DefaultCertChain, "").Get()
	tlsServerKey       = env.RegisterStringVar(bootstrap.IstioMetaPrefix+model.NodeMetadataTLSServerKey, constants.DefaultKey, "").Get()
	tlsServerRootCert  = env.RegisterStringVar(bootstrap.IstioMetaPrefix+model.NodeMetadataTLSServerRootCert, constants.DefaultRootCert, "").Get()

	tlsClientCertChain = env.RegisterStringVar(bootstrap.IstioMetaPrefix+model.NodeMetadataTLSClientCertChain, constants.DefaultCertChain, "").Get()
	tlsClientKey       = env.RegisterStringVar(bootstrap.IstioMetaPrefix+model.NodeMetadataTLSClientKey, constants.DefaultKey, "").Get()
	tlsClientRootCert  = env.RegisterStringVar(bootstrap.IstioMetaPrefix+model.NodeMetadataTLSClientRootCert, constants.DefaultRootCert, "").Get()
)

// Istio-agent requires recent K8S with JWT support. This is used to authenticate with control plane
// and get certificates.
const trustworthyJWTPath = "/var/run/secrets/tokens/istio-token"

type AgentConfig struct {
	ProxyIP          string
	Registry         serviceregistry.ServiceRegistry
	TrustDomain      string
	PilotIdentity    string
	MixerIdentity    string
	StatusPort       uint16
	ApplicationPorts []string

	// proxy config flags (named identically)
	ConfigPath               string
	ControlPlaneBootstrap    bool
	BinaryPath               string
	ServiceCluster           string
	DrainDuration            time.Duration
	ParentShutdownDuration   time.Duration
	DiscoveryAddress         string
	ZipkinAddress            string
	LightstepAddress         string
	LightstepAccessToken     string
	LightstepSecure          bool
	LightstepCacertPath      string
	DatadogAgentAddress      string
	ConnectTimeout           time.Duration
	StatsdUDPAddress         string
	EnvoyMetricsService      string
	EnvoyAccessLogService    string
	ProxyAdminPort           uint16
	ControlPlaneAuthPolicy   string
	ProxyLogLevel            string
	ProxyComponentLogLevel   string
	DNSRefreshRate           string
	Concurrency              int
	TemplateFile             string
	DisableInternalTelemetry bool
	TlsCertsToWatch          []string
}

var (
	Role           = &model.Proxy{}
	LoggingOptions = log.DefaultOptions()

	wg sync.WaitGroup

	instanceIPVar             = env.RegisterStringVar("INSTANCE_IP", "", "")
	podNameVar                = env.RegisterStringVar("POD_NAME", "", "")
	podNamespaceVar           = env.RegisterStringVar("POD_NAMESPACE", "", "")
	istioNamespaceVar         = env.RegisterStringVar("ISTIO_NAMESPACE", "", "")
	kubeAppProberNameVar      = env.RegisterStringVar(status.KubeAppProberEnvName, "", "")
	sdsEnabledVar             = env.RegisterBoolVar("SDS_ENABLED", false, "")
	sdsUdsPathVar             = env.RegisterStringVar("SDS_UDS_PATH", "unix:/var/run/sds/uds_path", "SDS address")
	stackdriverTracingEnabled = env.RegisterBoolVar("STACKDRIVER_TRACING_ENABLED", false, "If enabled, stackdriver will"+
		" get configured as the tracer.")
	stackdriverTracingDebug = env.RegisterBoolVar("STACKDRIVER_TRACING_DEBUG", false, "If set to true, "+
		"enables trace output to stdout")
	stackdriverTracingMaxNumberOfAnnotations = env.RegisterIntVar("STACKDRIVER_TRACING_MAX_NUMBER_OF_ANNOTATIONS", 200, "Sets the max"+
		" number of annotations for stackdriver")
	stackdriverTracingMaxNumberOfAttributes = env.RegisterIntVar("STACKDRIVER_TRACING_MAX_NUMBER_OF_ATTRIBUTES", 200, "Sets the max "+
		"number of attributes for stackdriver")
	stackdriverTracingMaxNumberOfMessageEvents = env.RegisterIntVar("STACKDRIVER_TRACING_MAX_NUMBER_OF_MESSAGE_EVENTS", 200, "Sets the "+
		"max number of message events for stackdriver")

	sdsUdsWaitTimeout = time.Minute

	// Indicates if any the remote services like AccessLogService, MetricsService have enabled tls.
	rsTLSEnabled bool

	ProxyConfig = mesh.DefaultProxyConfig()

	Opts *AgentConfig
)

// Start envoy.
// Typically called with parameters injected by istio injector and K8S, via env variables.
// In pilot-agent, the command was:
//   "proxy sidecar --domain validation-temp-ns.svc.cluster.local
//     --configPath /etc/istio/proxy
//     --binaryPath /usr/local/bin/envoy
//     --serviceCluster sleep.validation-temp-ns
//     --drainDuration45s
//     --parentShutdownDuration1m0s
//     --discoveryAddress istio-pilot.istio-system:15011
//     --zipkinAddress zipkin.istio-system:9411
//     --dnsRefreshRate 300s
//     --connectTimeout 10s
//     --proxyAdminPort 15000
//     --concurrency 2
//     --controlPlaneAuthPolicy MUTUAL_TLS
//     --statusPort 15020
//     --applicationPorts
//     --trust-domain=costin-istio.svc.id.goog"
//
// role can be 'sidecar' or 'router'
//
func Start(roleName string, opts *AgentConfig) error {
	if err := log.Configure(LoggingOptions); err != nil {
		return err
	}
	Opts = opts

	// Extract pod variables.
	podName := podNameVar.Get()
	podNamespace := podNamespaceVar.Get()
	podIP := net.ParseIP(instanceIPVar.Get()) // protobuf encoding of IP_ADDRESS type

	Role.Type = model.NodeType(roleName)
	if !model.IsApplicationNodeType(Role.Type) {
		log.Errorf("Invalid role Type: %#v", Role.Type)
		return fmt.Errorf("Invalid role Type: " + string(Role.Type))
	}

	//Do we need to get IP from the command line or environment?
	if len(opts.ProxyIP) != 0 {
		Role.IPAddresses = append(Role.IPAddresses, opts.ProxyIP)
	} else if podIP != nil {
		Role.IPAddresses = append(Role.IPAddresses, podIP.String())
	}

	// Obtain all the IPs from the node
	if ipAddrs, ok := proxy.GetPrivateIPs(context.Background()); ok {
		log.Infof("Obtained private IP %v", ipAddrs)
		Role.IPAddresses = append(Role.IPAddresses, ipAddrs...)
	}

	// No IP addresses provided, append 127.0.0.1 for ipv4 and ::1 for ipv6
	if len(Role.IPAddresses) == 0 {
		Role.IPAddresses = append(Role.IPAddresses, "127.0.0.1")
		Role.IPAddresses = append(Role.IPAddresses, "::1")
	}

	// Check if proxy runs in ipv4 or ipv6 environment to set Envoy's
	// operational parameters correctly.
	proxyIPv6 := isIPv6Proxy(Role.IPAddresses)
	if len(Role.ID) == 0 {
		if opts.Registry == serviceregistry.KubernetesRegistry {
			Role.ID = podName + "." + podNamespace
		} else if opts.Registry == serviceregistry.ConsulRegistry {
			Role.ID = Role.IPAddresses[0] + ".service.consul"
		} else {
			Role.ID = Role.IPAddresses[0]
		}
	}

	opts.TrustDomain = spiffe.DetermineTrustDomain(opts.TrustDomain, true)
	spiffe.SetTrustDomain(opts.TrustDomain)
	log.Infof("Proxy role: %#v", Role)

	opts.TlsCertsToWatch = []string{
		tlsServerCertChain, tlsServerKey, tlsServerRootCert,
		tlsClientCertChain, tlsClientKey, tlsClientRootCert,
	}

	// set all flags
	ProxyConfig.ConfigPath = opts.ConfigPath
	ProxyConfig.BinaryPath = opts.BinaryPath
	ProxyConfig.ServiceCluster = opts.ServiceCluster
	ProxyConfig.DrainDuration = types.DurationProto(opts.DrainDuration)
	ProxyConfig.ParentShutdownDuration = types.DurationProto(opts.ParentShutdownDuration)
	ProxyConfig.DiscoveryAddress = opts.DiscoveryAddress
	ProxyConfig.ConnectTimeout = types.DurationProto(opts.ConnectTimeout)
	ProxyConfig.StatsdUdpAddress = opts.StatsdUDPAddress
	if opts.EnvoyMetricsService != "" {
		if ms := fromJSON(opts.EnvoyMetricsService); ms != nil {
			ProxyConfig.EnvoyMetricsService = ms
			appendTLSCerts(ms)
		}
	}
	if opts.EnvoyAccessLogService != "" {
		if rs := fromJSON(opts.EnvoyAccessLogService); rs != nil {
			ProxyConfig.EnvoyAccessLogService = rs
			appendTLSCerts(rs)
		}
	}
	ProxyConfig.ProxyAdminPort = int32(opts.ProxyAdminPort)
	ProxyConfig.Concurrency = int32(opts.Concurrency)

	var pilotSAN []string
	controlPlaneAuthEnabled := false
	ns := ""
	switch opts.ControlPlaneAuthPolicy {
	case meshconfig.AuthenticationPolicy_NONE.String():
		ProxyConfig.ControlPlaneAuthPolicy = meshconfig.AuthenticationPolicy_NONE
	case meshconfig.AuthenticationPolicy_MUTUAL_TLS.String():
		controlPlaneAuthEnabled = true
		ProxyConfig.ControlPlaneAuthPolicy = meshconfig.AuthenticationPolicy_MUTUAL_TLS
		if opts.Registry == serviceregistry.KubernetesRegistry {
			partDiscoveryAddress := strings.Split(opts.DiscoveryAddress, ":")
			discoveryHostname := partDiscoveryAddress[0]
			parts := strings.Split(discoveryHostname, ".")
			if len(parts) == 1 {
				// namespace of pilot is not part of discovery address use
				// pod namespace e.g. istio-pilot:15005
				ns = podNamespace
			} else if len(parts) == 2 {
				// namespace is found in the discovery address
				// e.g. istio-pilot.istio-system:15005
				ns = parts[1]
			} else {
				// discovery address is a remote address. For remote clusters
				// only support the default config, or env variable
				ns = istioNamespaceVar.Get()
				if ns == "" {
					ns = constants.IstioSystemNamespace
				}
			}
		}
	}
	Role.DNSDomain = getDNSDomain(podNamespace, Role.DNSDomain)
	setSpiffeTrustDomain(podNamespace, Role.DNSDomain)

	// Obtain the Pilot and Mixer SANs. Used below to create a Envoy proxy.
	pilotSAN = getSAN(ns, envoyDiscovery.PilotSvcAccName, opts.PilotIdentity)
	log.Infof("PilotSAN %#v", pilotSAN)
	mixerSAN := getSAN(ns, envoyDiscovery.MixerSvcAccName, opts.MixerIdentity)
	log.Infof("MixerSAN %#v", mixerSAN)

	// resolve statsd address
	if ProxyConfig.StatsdUdpAddress != "" {
		addr, err := proxy.ResolveAddr(ProxyConfig.StatsdUdpAddress)
		if err != nil {
			// If istio-mixer.istio-system can't be resolved, skip generating the statsd config.
			// (instead of crashing). Mixer is optional.
			log.Warnf("resolve StatsdUdpAddress failed: %v", err)
			ProxyConfig.StatsdUdpAddress = ""
		} else {
			ProxyConfig.StatsdUdpAddress = addr
		}
	}

	// set tracing config
	if opts.LightstepAddress != "" {
		ProxyConfig.Tracing = &meshconfig.Tracing{
			Tracer: &meshconfig.Tracing_Lightstep_{
				Lightstep: &meshconfig.Tracing_Lightstep{
					Address:     opts.LightstepAddress,
					AccessToken: opts.LightstepAccessToken,
					Secure:      opts.LightstepSecure,
					CacertPath:  opts.LightstepCacertPath,
				},
			},
		}
	} else if opts.ZipkinAddress != "" {
		ProxyConfig.Tracing = &meshconfig.Tracing{
			Tracer: &meshconfig.Tracing_Zipkin_{
				Zipkin: &meshconfig.Tracing_Zipkin{
					Address: opts.ZipkinAddress,
				},
			},
		}
	} else if opts.DatadogAgentAddress != "" {
		ProxyConfig.Tracing = &meshconfig.Tracing{
			Tracer: &meshconfig.Tracing_Datadog_{
				Datadog: &meshconfig.Tracing_Datadog{
					Address: opts.DatadogAgentAddress,
				},
			},
		}
	} else if stackdriverTracingEnabled.Get() {
		ProxyConfig.Tracing = &meshconfig.Tracing{
			Tracer: &meshconfig.Tracing_Stackdriver_{
				Stackdriver: &meshconfig.Tracing_Stackdriver{
					Debug: stackdriverTracingDebug.Get(),
					MaxNumberOfAnnotations: &types.Int64Value{
						Value: int64(stackdriverTracingMaxNumberOfAnnotations.Get()),
					},
					MaxNumberOfAttributes: &types.Int64Value{
						Value: int64(stackdriverTracingMaxNumberOfAttributes.Get()),
					},
					MaxNumberOfMessageEvents: &types.Int64Value{
						Value: int64(stackdriverTracingMaxNumberOfMessageEvents.Get()),
					},
				},
			},
		}
	}

	if err := validation.ValidateProxyConfig(&ProxyConfig); err != nil {
		return err
	}

	if out, err := gogoprotomarshal.ToYAML(&ProxyConfig); err != nil {
		log.Infof("Failed to serialize to YAML: %v", err)
	} else {
		log.Infof("Effective config: %s", out)
	}

	sdsUDSPath := sdsUdsPathVar.Get()
	sdsEnabled, sdsTokenPath := detectSds(opts.ControlPlaneBootstrap, sdsUDSPath, trustworthyJWTPath)
	// dedupe cert paths so we don't set up 2 watchers for the same file:
	opts.TlsCertsToWatch = dedupeStrings(opts.TlsCertsToWatch)

	// Since Envoy needs the file-mounted certs for mTLS, we wait for them to become available
	// before starting it. Skip waiting cert if sds is enabled, otherwise it takes long time for
	// pod to start.
	if (controlPlaneAuthEnabled || rsTLSEnabled) && !sdsEnabled {
		log.Infof("Monitored certs: %#v", opts.TlsCertsToWatch)
		for _, cert := range opts.TlsCertsToWatch {
			waitForFile(cert, 2*time.Minute)
		}
	}

	// If control plane auth is not mTLS or global SDS flag is turned off, unset UDS path and token path
	// for control plane SDS.
	if !controlPlaneAuthEnabled || !sdsEnabled {
		sdsUDSPath = ""
		sdsTokenPath = ""
	}

	if opts.TemplateFile != "" && ProxyConfig.CustomConfigFile == "" {
		ProxyConfig.ProxyBootstrapTemplatePath = opts.TemplateFile
	}

	ctx, cancel := context.WithCancel(context.Background())
	// If a status port was provided, start handling status probes.
	if opts.StatusPort > 0 {
		parsedPorts, err := parseApplicationPorts()
		if err != nil {
			cancel()
			return err
		}
		localHostAddr := "127.0.0.1"
		if proxyIPv6 {
			localHostAddr = "[::1]"
		}
		prober := kubeAppProberNameVar.Get()
		statusServer, err := status.NewServer(status.Config{
			LocalHostAddr:      localHostAddr,
			AdminPort:          opts.ProxyAdminPort,
			StatusPort:         opts.StatusPort,
			ApplicationPorts:   parsedPorts,
			KubeAppHTTPProbers: prober,
			NodeType:           Role.Type,
		})
		if err != nil {
			cancel()
			return err
		}
		go waitForCompletion(ctx, statusServer.Run)
	}

	log.Infof("PilotSAN %#v", pilotSAN)

	envoyProxy := envoy.NewProxy(envoy.ProxyConfig{
		Config:              ProxyConfig,
		Node:                Role.ServiceNode(),
		LogLevel:            opts.ProxyLogLevel,
		ComponentLogLevel:   opts.ProxyComponentLogLevel,
		PilotSubjectAltName: pilotSAN,
		MixerSubjectAltName: mixerSAN,
		NodeIPs:             Role.IPAddresses,
		DNSRefreshRate:      opts.DNSRefreshRate,
		PodName:             podName,
		PodNamespace:        podNamespace,
		PodIP:               podIP,
		SDSUDSPath:          sdsUDSPath,
		SDSTokenPath:        sdsTokenPath,
		ControlPlaneAuth:    controlPlaneAuthEnabled,
		DisableReportCalls:  opts.DisableInternalTelemetry,
	})

	agent := envoy.NewAgent(envoyProxy, features.TerminationDrainDuration())

	watcher := envoy.NewWatcher(opts.TlsCertsToWatch, agent.Restart)

	go watcher.Run(ctx)

	// On SIGINT or SIGTERM, cancel the context, triggering a graceful shutdown
	go cmd.WaitSignalFunc(cancel)

	return agent.Run(ctx)

}

// dedupes the string array and also ignores the empty string.
func dedupeStrings(in []string) []string {
	stringMap := map[string]bool{}
	for _, c := range in {
		if len(c) > 0 {
			stringMap[c] = true
		}
	}
	unique := make([]string, 0)
	for c := range stringMap {
		unique = append(unique, c)
	}
	return unique
}

func waitForCompletion(ctx context.Context, fn func(context.Context)) {
	wg.Add(1)
	fn(ctx)
	wg.Done()
}

//explicitly setting the trustdomain so the pilot and mixer SAN will have same trustdomain
//and the initialization of the spiffe pkg isn't linked to generating pilot's SAN first
func setSpiffeTrustDomain(podNamespace string, domain string) {
	if Opts.ControlPlaneAuthPolicy == meshconfig.AuthenticationPolicy_MUTUAL_TLS.String() {
		pilotTrustDomain := Opts.TrustDomain
		if len(pilotTrustDomain) == 0 {
			if Opts.Registry == serviceregistry.KubernetesRegistry &&
				(domain == podNamespace+".svc.cluster.local" || domain == "") {
				pilotTrustDomain = "cluster.local"
			} else if Opts.Registry == serviceregistry.ConsulRegistry &&
				(domain == "service.consul" || domain == "") {
				pilotTrustDomain = ""
			} else {
				pilotTrustDomain = domain
			}
		}
		spiffe.SetTrustDomain(pilotTrustDomain)
	}

}

func getSAN(ns string, defaultSA string, overrideIdentity string) []string {
	var san []string
	if Opts.ControlPlaneAuthPolicy == meshconfig.AuthenticationPolicy_MUTUAL_TLS.String() {

		if overrideIdentity == "" {
			san = append(san, envoyDiscovery.GetSAN(ns, defaultSA))
		} else {
			san = append(san, envoyDiscovery.GetSAN("", overrideIdentity))
		}
	}
	return san
}

func getDNSDomain(podNamespace, domain string) string {
	if len(domain) == 0 {
		if Opts.Registry == serviceregistry.KubernetesRegistry {
			domain = podNamespace + ".svc.cluster.local"
		} else if Opts.Registry == serviceregistry.ConsulRegistry {
			domain = "service.consul"
		} else {
			domain = ""
		}
	}
	return domain
}

// detectSds checks if the SDS address (when it is UDS) and JWT paths are present.
func detectSds(controlPlaneBootstrap bool, sdsAddress, trustworthyJWTPath string) (bool, string) {
	if !sdsEnabledVar.Get() {
		return false, ""
	}

	if len(sdsAddress) == 0 {
		return false, ""
	}

	if _, err := os.Stat(trustworthyJWTPath); err != nil {
		return false, ""
	}

	// sdsAddress will not be empty when sdsAddress is a UDS address.
	udsPath := ""
	if strings.HasPrefix(sdsAddress, "unix:") {
		udsPath = strings.TrimPrefix(sdsAddress, "unix:")
		if len(udsPath) == 0 {
			// If sdsAddress is "unix:", it is invalid, return false.
			return false, ""
		}
	} else {
		return true, trustworthyJWTPath
	}

	if !controlPlaneBootstrap {
		// workload sidecar
		// treat sds as disabled if uds path isn't set.
		if _, err := os.Stat(udsPath); err != nil {
			return false, ""
		}

		return true, trustworthyJWTPath
	}

	// controlplane components like pilot/mixer/galley have sidecar
	// they start almost same time as sds server; wait since there is a chance
	// when pilot-agent start, the uds file doesn't exist.
	if !waitForFile(udsPath, sdsUdsWaitTimeout) {
		return false, ""
	}

	return true, trustworthyJWTPath
}

func parseApplicationPorts() ([]uint16, error) {
	parsedPorts := make([]uint16, 0, len(Opts.ApplicationPorts))
	for _, port := range Opts.ApplicationPorts {
		port := strings.TrimSpace(port)
		if len(port) > 0 {
			parsedPort, err := strconv.ParseUint(port, 10, 16)
			if err != nil {
				return nil, err
			}
			parsedPorts = append(parsedPorts, uint16(parsedPort))
		}
	}
	return parsedPorts, nil
}

func TimeDuration(dur *types.Duration) time.Duration {
	out, err := types.DurationFromProto(dur)
	if err != nil {
		log.Warna(err)
	}
	return out
}

func fromJSON(j string) *meshconfig.RemoteService {
	var m meshconfig.RemoteService
	err := jsonpb.UnmarshalString(j, &m)
	if err != nil {
		log.Warnf("Unable to unmarshal %s: %v", j, err)
		return nil
	}

	return &m
}

func appendTLSCerts(rs *meshconfig.RemoteService) {
	if rs.TlsSettings == nil {
		return
	}
	if rs.TlsSettings.Mode == networking.TLSSettings_DISABLE {
		return
	}
	rsTLSEnabled = true
	Opts.TlsCertsToWatch = append(Opts.TlsCertsToWatch, rs.TlsSettings.CaCertificates, rs.TlsSettings.ClientCertificate,
		rs.TlsSettings.PrivateKey)
}

func waitForFile(fname string, maxWait time.Duration) bool {
	log.Infof("waiting %v for %s", maxWait, fname)

	logDelay := 1 * time.Second
	nextLog := time.Now().Add(logDelay)
	endWait := time.Now().Add(maxWait)

	for {
		_, err := os.Stat(fname)
		if err == nil {
			return true
		}
		if !os.IsNotExist(err) { // another error (e.g., permission) - likely no point in waiting longer
			log.Errora("error while waiting for file", err.Error())
			return false
		}

		now := time.Now()
		if now.After(endWait) {
			log.Warna("file still not available after", maxWait)
			return false
		}
		if now.After(nextLog) {
			log.Infof("waiting for file")
			logDelay *= 2
			nextLog.Add(logDelay)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// TODO: get the config and bootstrap from istiod, by passing the env

// isIPv6Proxy check the addresses slice and returns true for a valid IPv6 address
// for all other cases it returns false
func isIPv6Proxy(ipAddrs []string) bool {
	for i := 0; i < len(ipAddrs); i++ {
		addr := net.ParseIP(ipAddrs[i])
		if addr == nil {
			// Should not happen, invalid IP in proxy's IPAddresses slice should have been caught earlier,
			// skip it to prevent a panic.
			continue
		}
		if addr.To4() != nil {
			return false
		}
	}
	return true
}
