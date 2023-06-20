package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/maxmoehl/harald"
	"golang.org/x/exp/slog"
	"gopkg.in/yaml.v3"
)

func main() {
	err := Main()
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		os.Exit(1)
	}
}

func Main() error {
	r, err := os.Open(os.Args[1])
	if err != nil {
		return fmt.Errorf("open envoy config: %w", err)
	}

	var envoyConfig EnvoyConfig
	err = yaml.NewDecoder(r).Decode(&envoyConfig)
	if err != nil {
		return fmt.Errorf("decode envoy config: %w", err)
	}

	configVersion := 2
	haraldConfig := harald.Config{
		Version: harald.Version{
			Version: &configVersion,
		},
		LogLevel:        slog.LevelDebug,
		Rules:           make(map[string]harald.ForwardRule),
		EnableListeners: true,
	}
	for _, listener := range envoyConfig.StaticResources.Listeners {
		rule := harald.ForwardRule{
			TLS: &harald.TLS{},
		}

		rule.Connect, err = getClusterNetConf(listener.FilterChains[0].Filters[0].TypedConfig.Cluster, envoyConfig)
		if err != nil {
			return fmt.Errorf("build rule: listener '%s': %w", listener.Name, err)
		}

		addr := listener.Address.SocketAddress
		rule.Listen.Network = "tcp"
		rule.Listen.Address = fmt.Sprintf("%s:%d", addr.Address, addr.Port)

		rule.TLS.Certificate, rule.TLS.Key, err = loadServerCertificate(listener.FilterChains[0].TransportSocket.TypedConfig.CommonTlsContext.ServerCertConfig[0].SdsConfig.Path)
		if err != nil {
			return fmt.Errorf("load server certificate: %w", err)
		}

		caPath := listener.FilterChains[0].TransportSocket.TypedConfig.CommonTlsContext.ClientCaConfig.SdsConfig.Path
		if caPath != "" {
			rule.TLS.ClientCAs, err = loadServerCa(caPath)
			if err != nil {
				return fmt.Errorf("load server certificate: %w", err)
			}
		}

		alpn := listener.FilterChains[0].TransportSocket.TypedConfig.CommonTlsContext.AlpnProtocols
		if len(alpn) > 0 {
			rule.TLS.ApplicationProtocols = strings.Split(alpn[0], ",")
		}

		haraldConfig.Rules[listener.Name] = rule
	}

	return yaml.NewEncoder(os.Stdout).Encode(haraldConfig)
}

func getClusterNetConf(name string, c EnvoyConfig) (harald.NetConf, error) {
	for _, cluster := range c.StaticResources.Clusters {
		if cluster.LoadAssignment.Name == name {
			addr := cluster.LoadAssignment.Endpoints[0].LbEndpoints[0].Endpoint.Address.SocketAddress

			return harald.NetConf{
				Network: "tcp",
				Address: fmt.Sprintf("%s:%d", addr.Address, addr.Port),
			}, nil
		}
	}
	return harald.NetConf{}, fmt.Errorf("cluster '%s': not found", name)
}

func loadServerCertificate(file string) (cert string, key string, err error) {
	r, err := os.Open(file)
	if err != nil {
		return "", "", fmt.Errorf("open server cert file: %w", err)
	}

	var c ServerCertificateConfig
	err = yaml.NewDecoder(r).Decode(&c)
	if err != nil {
		return "", "", fmt.Errorf("decode server cert file: %w", err)
	}

	return c.Resources[0].TlsCertificate.CertificateChain.InlineString, c.Resources[0].TlsCertificate.PrivateKey.InlineString, nil
}

func loadServerCa(file string) (cert string, err error) {
	r, err := os.Open(file)
	if err != nil {
		return "", fmt.Errorf("open server cert file: %w", err)
	}

	var c ValidationContext
	err = yaml.NewDecoder(r).Decode(&c)
	if err != nil {
		return "", fmt.Errorf("decode ca file: %w", err)
	}

	return c.Resources[0].ValidationContext.TrustedCa.InlineString, nil
}

type EnvoyConfig struct {
	StaticResources struct {
		Clusters []struct {
			LoadAssignment struct {
				Name      string `yaml:"cluster_name"`
				Endpoints []struct {
					LbEndpoints []struct {
						Endpoint struct {
							Address struct {
								SocketAddress struct {
									Address string `yaml:"address"`
									Port    int    `yaml:"port_value"`
								} `yaml:"socket_address"`
							} `yaml:"address"`
						} `yaml:"endpoint"`
					} `yaml:"lb_endpoints"`
				} `yaml:"endpoints"`
			} `yaml:"load_assignment"`
		} `yaml:"clusters"`
		Listeners []struct {
			Name    string `yaml:"name"`
			Address struct {
				SocketAddress struct {
					Address string `yaml:"address"`
					Port    int    `yaml:"port_value"`
				} `yaml:"socket_address"`
			} `yaml:"address"`
			FilterChains []struct {
				Filters []struct {
					TypedConfig struct {
						Cluster string `yaml:"cluster"`
					} `yaml:"typed_config"`
				} `yaml:"filters"`
				TransportSocket struct {
					TypedConfig struct {
						CommonTlsContext struct {
							AlpnProtocols    []string `yaml:"alpn_protocols"`
							ServerCertConfig []struct {
								Name      string `yaml:"name"`
								SdsConfig struct {
									Path string `yaml:"path"`
								} `yaml:"sds_config"`
							} `yaml:"tls_certificate_sds_secret_configs"`
							ClientCaConfig struct {
								SdsConfig struct {
									Path string `yaml:"path"`
								} `yaml:"sds_config"`
							} `yaml:"validation_context_sds_secret_config"`
						} `yaml:"common_tls_context"`
					} `yaml:"typed_config"`
				} `yaml:"transport_socket"`
			} `yaml:"filter_chains"`
		} `yaml:"listeners"`
	} `yaml:"static_resources"`
}

type ServerCertificateConfig struct {
	Resources []struct {
		Name           string `yaml:"name"`
		TlsCertificate struct {
			CertificateChain struct {
				InlineString string `yaml:"inline_string"`
			} `yaml:"certificate_chain"`
			PrivateKey struct {
				InlineString string `yaml:"inline_string"`
			} `yaml:"private_key"`
		} `yaml:"tls_certificate"`
	} `yaml:"resources"`
}

type ValidationContext struct {
	Resources []struct {
		Name              string `yaml:"name"`
		ValidationContext struct {
			TrustedCa struct {
				InlineString string `yaml:"inline_string"`
			} `yaml:"trusted_ca"`
			MatchSubjectAltNames []struct {
				Exact string `yaml:"exact"`
			} `yaml:"private_key"`
		} `yaml:"validation_context"`
	} `yaml:"resources"`
}
