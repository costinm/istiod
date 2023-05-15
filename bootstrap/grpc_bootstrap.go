// Copyright Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bootstrap

//import (
//	"encoding/json"
//	"fmt"
//	"io/ioutil"
//	"os"
//	"path"
//	"strings"
//	"unicode/utf8"
//
//	core "github.com/costinm/istiod/gen/grpc-go/xds"
//	"google.golang.org/protobuf/runtime/protoimpl"
//	"google.golang.org/protobuf/types/known/structpb"
//
//	log "google.golang.org/grpc/grpclog"
//)
//
//// Extracted from istio/istio/pkg/istio-agent/grpcxds
//// - removed deps to istio internal packages
//
//// If a sidecar is present:
//// - "unix:///etc/istio/proxy/XDS" using agent
//// - istiod.istio-system.svc:15010 for plaintext
//
//const (
//	ServerListenerNamePrefix = "xds.istio.io/grpc/lds/inbound/"
//	// ServerListenerNameTemplate for the name of the Listener resource to subscribe to for a gRPC
//	// server. If the token `%s` is present in the string, all instances of the
//	// token will be replaced with the server's listening "IP:port" (e.g.,
//	// "0.0.0.0:8080", "[::]:8080").
//	ServerListenerNameTemplate = ServerListenerNamePrefix + "%s"
//)
//
//// Bootstrap contains the general structure of what's expected by GRPC's XDS implementation.
//// See https://github.com/grpc/grpc-go/blob/master/xds/internal/xdsclient/bootstrap/bootstrap.go
//// TODO use structs from gRPC lib if created/exported
//type Bootstrap struct {
//	XDSServers                 []XdsServer                    `json:"xds_servers,omitempty"`
//	Node                       *core.Node                     `json:"node,omitempty"`
//	CertProviders              map[string]CertificateProvider `json:"certificate_providers,omitempty"`
//	ServerListenerNameTemplate string                         `json:"server_listener_resource_name_template,omitempty"`
//}
//
//type ChannelCreds struct {
//	Type   string      `json:"type,omitempty"`
//	Config interface{} `json:"config,omitempty"`
//}
//
//type XdsServer struct {
//	ServerURI      string         `json:"server_uri,omitempty"`
//	ChannelCreds   []ChannelCreds `json:"channel_creds,omitempty"`
//	ServerFeatures []string       `json:"server_features,omitempty"`
//}
//
//type CertificateProvider struct {
//	PluginName string      `json:"plugin_name,omitempty"`
//	Config     interface{} `json:"config,omitempty"`
//}
//
//func (cp *CertificateProvider) UnmarshalJSON(data []byte) error {
//	var dat map[string]*json.RawMessage
//	if err := json.Unmarshal(data, &dat); err != nil {
//		return err
//	}
//	*cp = CertificateProvider{}
//
//	if pluginNameVal, ok := dat["plugin_name"]; ok {
//		if err := json.Unmarshal(*pluginNameVal, &cp.PluginName); err != nil {
//			log.Warningf("failed parsing plugin_name in certificate_provider: %v", err)
//		}
//	} else {
//		log.Warningf("did not find plugin_name in certificate_provider")
//	}
//
//	if configVal, ok := dat["config"]; ok {
//		var err error
//		switch cp.PluginName {
//		case FileWatcherCertProviderName:
//			config := FileWatcherCertProviderConfig{}
//			err = json.Unmarshal(*configVal, &config)
//			cp.Config = config
//		default:
//			config := FileWatcherCertProviderConfig{}
//			err = json.Unmarshal(*configVal, &config)
//			cp.Config = config
//		}
//		if err != nil {
//			log.Warningf("failed parsing config in certificate_provider: %v", err)
//		}
//	} else {
//		log.Warningf("did not find config in certificate_provider")
//	}
//
//	return nil
//}
//
//const FileWatcherCertProviderName = "file_watcher"
//
//type FileWatcherCertProviderConfig struct {
//	CertificateFile   string `json:"certificate_file,omitempty"`
//	PrivateKeyFile    string `json:"private_key_file,omitempty"`
//	CACertificateFile string `json:"ca_certificate_file,omitempty"`
//	RefreshDuration   string `json:"refresh_interval,omitempty"`
//}
//
//func (c *FileWatcherCertProviderConfig) FilePaths() []string {
//	return []string{c.CertificateFile, c.PrivateKeyFile, c.CACertificateFile}
//}
//
//// FileWatcherProvider returns the FileWatcherCertProviderConfig if one exists in CertProviders
//func (b *Bootstrap) FileWatcherProvider() *FileWatcherCertProviderConfig {
//	if b == nil || b.CertProviders == nil {
//		return nil
//	}
//	for _, provider := range b.CertProviders {
//		if provider.PluginName == FileWatcherCertProviderName {
//			cfg, ok := provider.Config.(FileWatcherCertProviderConfig)
//			if !ok {
//				return nil
//			}
//			return &cfg
//		}
//	}
//	return nil
//}
//
//// LoadBootstrap loads a Bootstrap from the given file path.
//func LoadBootstrap(file string) (*Bootstrap, error) {
//	data, err := os.ReadFile(file)
//	if err != nil {
//		return nil, err
//	}
//	b := &Bootstrap{}
//	if err := json.Unmarshal(data, b); err != nil {
//		return nil, err
//	}
//	return b, err
//}
//
//// GenerateBootstrap generates the bootstrap structure for gRPC XDS integration.
//// This is used for 'agentless' - but should also work if an agent is used (or some other provider handles the XDS proxy).
//func GenerateBootstrap(opts *GenerateBootstrapOptions) (*Bootstrap, error) {
//	if opts == nil {
//		opts = &GenerateBootstrapOptions{}
//		if _, err := os.Stat(defaultXDSProxy); os.IsNotExist(err) {
//			// TODO: Detect XDS address
//			opts.DiscoveryAddress = "localhost:15010"
//		}
//	}
//
//	xdsMeta, err := extractMeta(opts.NodeMetadata)
//	if err != nil {
//		return nil, fmt.Errorf("failed extracting xds metadata: %v", err)
//	}
//
//	bootstrap := &Bootstrap{
//		XDSServers: []XdsServer{{
//			// connect locally via agent
//			ServerFeatures: []string{"xds_v3"},
//		}},
//		Node: &core.Node{
//			Id:       opts.ID,
//			Locality: opts.Locality,
//			Metadata: xdsMeta,
//		},
//		ServerListenerNameTemplate: ServerListenerNameTemplate,
//	}
//
//	// TODO direct to CP should use secure channel (most likely JWT + TLS, but possibly allow mTLS)
//	serverURI := opts.DiscoveryAddress
//	if serverURI == "" {
//		if opts.XdsUdsPath != "" {
//			serverURI = fmt.Sprintf("unix:///%s", opts.XdsUdsPath)
//			bootstrap.XDSServers[0].ChannelCreds = []ChannelCreds{{Type: "insecure"}}
//		} else if _, err := os.Stat(defaultXDSProxy); !os.IsNotExist(err) {
//			serverURI = "unix://" + defaultXDSProxy
//			bootstrap.XDSServers[0].ChannelCreds = []ChannelCreds{{Type: "insecure"}}
//		}
//	} else {
//		// TODO: add support to Istiod to support JWTs from google
//		if strings.Contains(serverURI, ":15010") {
//			bootstrap.XDSServers[0].ChannelCreds = []ChannelCreds{{Type: "insecure"}}
//		} else {
//			bootstrap.XDSServers[0].ChannelCreds = []ChannelCreds{{Type: "mtls"}}
//		}
//	}
//	bootstrap.XDSServers[0].ServerURI = serverURI
//
//	InitCerts(bootstrap, opts)
//	return bootstrap, err
//}
//
//func InitCerts(bootstrap *Bootstrap, opts *GenerateBootstrapOptions) {
//	if false {
//		// If this code is called, we have registered the custom xds auth and provider.
//		bootstrap.CertProviders = map[string]CertificateProvider{
//			"default": {
//				PluginName: "certs", // Same as Name()
//			},
//		}
//		return
//	}
//	if opts.CertDir != "" {
//		bootstrap.CertProviders = map[string]CertificateProvider{
//			"default": {
//				PluginName: "file_watcher",
//				Config: FileWatcherCertProviderConfig{
//					PrivateKeyFile:    path.Join(opts.CertDir, "key.pem"),
//					CertificateFile:   path.Join(opts.CertDir, "cert-chain.pem"),
//					CACertificateFile: path.Join(opts.CertDir, "root-cert.pem"),
//					RefreshDuration:   "600s",
//				},
//			},
//		}
//		return
//	}
//	base := "/var/run/secrets/workload-spiffe-credentials/"
//	if _, err := os.Stat(base + "ca_certificates.pem"); !os.IsNotExist(err) {
//		bootstrap.CertProviders = map[string]CertificateProvider{
//			"default": {
//				PluginName: "file_watcher",
//				Config: FileWatcherCertProviderConfig{
//					PrivateKeyFile:    path.Join(base, "private_key.pem"),
//					CertificateFile:   path.Join(base, "certificates.pem"),
//					CACertificateFile: path.Join(base, "ca_certificates.pem"),
//					RefreshDuration:   "600s",
//				},
//			},
//		}
//		return
//	}
//
//	// Didn't find platform files, attempt in-memory
//
//}
//
//func extractMeta(v map[string]interface{}) (*core.Struct, error) {
//	x := &core.Struct{Fields: make(map[string]*core.Value, len(v))}
//	for k, v := range v {
//		if !utf8.ValidString(k) {
//			return nil, protoimpl.X.NewError("invalid UTF-8 in string: %q", k)
//		}
//		var err error
//		x.Fields[k], err = NewValue(v)
//		if err != nil {
//			return nil, err
//		}
//	}
//	xdsMeta, err := structpb.NewStruct(rawMeta)
//	if err != nil {
//		return nil, err
//	}
//	return xdsMeta, nil
//}
//
//// GenerateBootstrapFile generates and writes atomically as JSON to the given file path.
//func GenerateBootstrapFile(opts *GenerateBootstrapOptions, path string) error {
//	bootstrap, err := GenerateBootstrap(opts)
//	if err != nil {
//		return err
//	}
//	jsonData, err := json.MarshalIndent(bootstrap, "", "  ")
//	if err != nil {
//		return err
//	}
//	if err := ioutil.WriteFile(path, jsonData, os.FileMode(0o644)); err != nil {
//		return fmt.Errorf("failed writing to %s: %v", path, err)
//	}
//	return nil
//}
//
////func Generate(opts *GenerateBootstrapOptions) error {
////	bootF := os.Getenv("GRPC_XDS_BOOTSTRAP")
////	if bootF == "" {
////		return errors.New("missing GRPC_XDS_BOOTSTRAP")
////	}
////
////	if _, err := os.Stat(bootF); os.IsNotExist(err) {
////		// TODO: write the bootstrap file.
////		GenerateBootstrapFile(opts, bootF)
////	}
////	return nil
////}
