package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/bugfan/ang-admin/cluster"
	"github.com/bugfan/ang-admin/entity"
	"github.com/bugfan/ang-admin/models"
)

func buildCertMap() map[string]entity.CertConfig {
	engine := models.GetEngine()
	if engine == nil {
		return nil
	}
	var certs []models.Certificate
	err := engine.Find(&certs)
	if err != nil {
		log.Printf("buildCertMap error: %v\n", err)
		return nil
	}

	certMap := make(map[string]entity.CertConfig)
	for _, item := range certs {
		certId := item.CertId
		if certId == "" {
			certId = fmt.Sprintf("id-%d", item.Id)
		}
		certMap[certId] = entity.CertConfig{
			Type: item.Type,
			Data: entity.CertData{
				Key:  item.KeyContent,
				Cert: item.CertContent,
			},
		}
	}
	return certMap
}

func buildTunnelMaps() (map[string]entity.TunnelConfig, map[string]entity.TunnelConfig) {
	engine := models.GetEngine()
	if engine == nil {
		return nil, nil
	}
	var tunnels []models.Tunnel
	err := engine.Find(&tunnels)
	if err != nil {
		log.Printf("buildTunnelMaps error: %v\n", err)
		return nil, nil
	}

	tlsMap := make(map[string]entity.TunnelConfig)
	quicMap := make(map[string]entity.TunnelConfig)

	for _, item := range tunnels {
		keyStr := strconv.FormatInt(item.Id, 10)
		cfg := entity.TunnelConfig{
			Port:        item.Port,
			Certificate: item.Certificate,
			Auth:        item.Auth,
		}

		tType := strings.ToUpper(strings.TrimSpace(item.Type))
		if tType == "TLS-TUNNEL" || tType == "TLS" {
			tlsMap[keyStr] = cfg
		} else if tType == "QUIC-TUNNEL" || tType == "QUIC" {
			quicMap[keyStr] = cfg
		} else {
			tlsMap[keyStr] = cfg
		}
	}
	return tlsMap, quicMap
}

func buildRulesDBMap() map[string]models.Rule {
	engine := models.GetEngine()
	if engine == nil {
		return nil
	}
	var rules []models.Rule
	err := engine.Find(&rules)
	if err != nil {
		log.Printf("buildRulesDBMap error: %v\n", err)
		return nil
	}

	rulesMap := make(map[string]models.Rule)
	for _, r := range rules {
		if r.Name != "" {
			rulesMap[r.Name] = r
		}
		rulesMap[strconv.FormatInt(r.Id, 10)] = r
	}
	return rulesMap
}

func buildDNSMap(rulesMap map[string]models.Rule) map[string]entity.DNSConfig {
	engine := models.GetEngine()
	if engine == nil {
		return nil
	}
	var dnsList []models.DnsProxy
	err := engine.Find(&dnsList)
	if err != nil {
		log.Printf("buildDNSMap error: %v\n", err)
		return nil
	}

	dnsMap := make(map[string]entity.DNSConfig)
	for _, item := range dnsList {
		keyStr := strconv.FormatInt(item.Id, 10)

		// Parse Hosts
		var hosts entity.DNSHosts
		if item.HostsJSON != "" {
			_ = json.Unmarshal([]byte(item.HostsJSON), &hosts)
		}

		// Parse Rules (Rule Set expansion)
		var ruleConfigs []entity.RuleConfig
		if item.Rules != "" {
			var ruleNames []string
			_ = json.Unmarshal([]byte(item.Rules), &ruleNames)
			for _, rName := range ruleNames {
				rName = strings.TrimSpace(rName)
				if dbRule, exists := rulesMap[rName]; exists {
					if dbRule.Items != "" {
						var items []entity.RuleConfig
						if err := json.Unmarshal([]byte(dbRule.Items), &items); err == nil {
							ruleConfigs = append(ruleConfigs, items...)
						}
					}
				} else if rName == "ip_matcher" {
					ruleConfigs = append(ruleConfigs, entity.RuleConfig{
						Matcher: entity.MatcherConfig{
							Name: "ip_matcher",
							Config: map[string]interface{}{
								"Address": []string{"121.0.0.1"},
							},
						},
						Action: entity.ActionConfig{
							Name: "reset_conn_action",
							Config: map[string]interface{}{
								"Content": "reset you",
							},
						},
					})
				}
			}
		}

		// Parse Backend
		var backend entity.DNSBackend
		if item.TunnelId != "" {
			backend.Tunnel = &entity.BackendTunnel{
				Type:  item.TunnelType,
				ID:    item.TunnelId,
				Token: item.TunnelToken,
			}
		}

		if item.UpstreamServers != "" {
			var servers []entity.UpstreamServer
			_ = json.Unmarshal([]byte(item.UpstreamServers), &servers)
			if len(servers) > 0 {
				method := item.UpstreamMethod
				if method == "" {
					method = "round_robin"
				}
				backend.Upstream = &entity.UpstreamConfig{
					Method: method,
					Data: entity.UpstreamData{
						Servers: servers,
					},
				}
			}
		}

		dnsMap[keyStr] = entity.DNSConfig{
			Address: item.Address,
			Port:    item.Port,
			Rule:    ruleConfigs,
			Hosts:   &hosts,
			Backend: &backend,
		}
	}
	return dnsMap
}

// buildRootCAConfig converts a PEM string to an entity.RootCAConfig pointer.
// Returns nil when pem is blank so that "RootCA" is omitted from JSON output entirely,
// which prevents the ang engine from replacing its system trust store with an empty pool.
func buildRootCAConfig(pem string) *entity.RootCAConfig {
	pem = strings.TrimSpace(pem)
	if pem == "" {
		return nil
	}
	return &entity.RootCAConfig{PEM: pem}
}

func buildHTTPMap(rulesMap map[string]models.Rule) map[string]entity.HTTPConfig {
	engine := models.GetEngine()
	if engine == nil {
		return nil
	}
	var httpList []models.HttpProxy
	err := engine.Find(&httpList)
	if err != nil {
		log.Printf("buildHTTPMap error: %v\n", err)
		return nil
	}

	var wildcardRootDomain string
	for _, p := range httpList {
		if strings.HasPrefix(p.Hostname, "*.") {
			wildcardRootDomain = strings.TrimPrefix(p.Hostname, "*.")
			break
		}
	}

	httpMap := make(map[string]entity.HTTPConfig)
	for _, item := range httpList {
		keyStr := strconv.FormatInt(item.Id, 10)

		// Parse ProxyHeaders
		var proxyHeaders []string
		if item.ProxyHeaders != "" {
			_ = json.Unmarshal([]byte(item.ProxyHeaders), &proxyHeaders)
		}

		// Parse Rules
		var ruleConfigs []entity.RuleConfig
		if item.Rules != "" {
			var ruleNames []string
			_ = json.Unmarshal([]byte(item.Rules), &ruleNames)
			for _, rName := range ruleNames {
				rName = strings.TrimSpace(rName)
				if dbRule, exists := rulesMap[rName]; exists {
					if dbRule.Items != "" {
						var items []entity.RuleConfig
						if err := json.Unmarshal([]byte(dbRule.Items), &items); err == nil {
							for i := range items {
								if items[i].Action.Name == "auth_portal_action" {
									if cfgMap, ok := items[i].Action.Config.(map[string]interface{}); ok {
										if cd, _ := cfgMap["cookie_domain"].(string); cd == "" && wildcardRootDomain != "" {
											if strings.HasSuffix(item.Hostname, wildcardRootDomain) {
												cfgMap["cookie_domain"] = "." + wildcardRootDomain
											}
										}
									}
								}
							}
							ruleConfigs = append(ruleConfigs, items...)
						}
					}
				}
			}
		}

		// Automatically synthesize and append subdomain_webvpn_action rule if active WebVPN sites exist on legacy HttpProxy
		var svcCount int64
		if engine != nil {
			svcCount, _ = engine.Count(new(models.WebvpnService))
		}
		if svcCount == 0 {
			var vpnSites []models.WebvpnSite
			_ = engine.Where("http_proxy_id = ? AND status = 1", item.Id).Find(&vpnSites)
			if len(vpnSites) > 0 {
				rootDomain := strings.TrimPrefix(item.Hostname, "*.")
				vpnActionSites := make(map[string]interface{})
				for _, vs := range vpnSites {
					u, err := url.Parse(vs.TargetURL)
					if err != nil {
						continue
					}
					targetHost := u.Hostname()
					targetPort := u.Port()
					schemePrefix := "s-"
					if u.Scheme == "http" {
						schemePrefix = "c-"
						if targetPort == "" {
							targetPort = "80"
						}
					} else {
						if targetPort == "" {
							targetPort = "443"
						}
					}

					hostMap := make(map[string]string)
					wildcardMap := make(map[string]string)

					// 1. Primary target host
					dashedTarget := strings.ReplaceAll(strings.ReplaceAll(targetHost, "-", "--"), ".", "-")
					mainVpnHost := fmt.Sprintf("%s%s-%s.%s", schemePrefix, dashedTarget, targetPort, rootDomain)
					hostMap[u.Host] = mainVpnHost
					if u.Port() != "" {
						hostMap[targetHost] = mainVpnHost
					}

					// 2. Related domain names from vs.Hosts
					for _, line := range strings.Split(vs.Hosts, "\n") {
						domain := strings.TrimSpace(line)
						if domain == "" {
							continue
						}
						if strings.Contains(domain, "*") {
							wildcardMap[domain] = ""
						} else {
							dashed := strings.ReplaceAll(strings.ReplaceAll(domain, "-", "--"), ".", "-")
							relVpnHost := fmt.Sprintf("%s%s-%s.%s", schemePrefix, dashed, targetPort, rootDomain)
							hostMap[domain] = relVpnHost
						}
					}

					isProt := vs.IsProtected == 1
					var groupIds []int64
					if isProt && vs.AllowedGroupIds != "" {
						_ = json.Unmarshal([]byte(vs.AllowedGroupIds), &groupIds)
					}

					replaceMap := make(map[string]string)
					if vs.Replace != "" {
						_ = json.Unmarshal([]byte(vs.Replace), &replaceMap)
					}

					vpnActionSites[vs.Prefix] = map[string]interface{}{
						"name":              vs.Name,
						"protected":         isProt,
						"allowed_group_ids": groupIds,
						"host":              hostMap,
						"wildcard":          wildcardMap,
						"replace":           replaceMap,
					}
				}

				if len(vpnActionSites) > 0 {
					loginURL := discoverAuthLoginURL(httpList, rulesMap)

					ruleConfigs = append(ruleConfigs, entity.RuleConfig{
						Matcher: entity.MatcherConfig{
							Name:   "always_true_matcher",
							Config: map[string]interface{}{},
						},
						Action: entity.ActionConfig{
							Name: "subdomain_webvpn_action",
							Config: map[string]interface{}{
								"Sites":        vpnActionSites,
								"LoginURL":     loginURL,
								"CookieDomain": "." + rootDomain,
								"Fallback":     "404",
							},
						},
					})
				}
			}
		}

		// Parse Backend Locations
		var rawLocations []entity.HTTPLocation
		var locations []entity.HTTPLocation
		if item.LocationJSON != "" {
			if err := json.Unmarshal([]byte(item.LocationJSON), &rawLocations); err == nil {
				for _, loc := range rawLocations {
					uType := loc.Upstream.Type
					if uType == "root" || uType == "alias" {
						// Clean legacy fields if any
						var dir string
						if mData, ok := loc.Upstream.Data.(map[string]interface{}); ok {
							if d, exists := mData["Dir"].(string); exists {
								dir = d
							}
						}
						if dir == "" {
							dir = "./static"
						}
						loc.Upstream.Data = map[string]interface{}{
							"Dir": dir,
						}
					} else {
						// proxy_pass
						var method string
						var servers interface{}
						if mData, ok := loc.Upstream.Data.(map[string]interface{}); ok {
							if m, exists := mData["Method"].(string); exists {
								method = m
							}
							if s, exists := mData["Servers"]; exists {
								servers = s
							}
						}
						if method == "" {
							method = "round_robin"
						}
						loc.Upstream.Type = "proxy_pass"
						loc.Upstream.Data = map[string]interface{}{
							"Method":  method,
							"Servers": servers,
						}
					}
					locations = append(locations, loc)
				}
			}
		}

		// Parse Backend Tunnel
		var backendTunnel *entity.BackendTunnel
		if item.TunnelId != "" {
			backendTunnel = &entity.BackendTunnel{
				Type:  item.TunnelType,
				ID:    item.TunnelId,
				Token: item.TunnelToken,
			}
		}

		var dnsResolver []string
		if item.DNSResolver != "" {
			if err := json.Unmarshal([]byte(item.DNSResolver), &dnsResolver); err != nil {
				// Fallback for old single string data
				dnsResolver = []string{item.DNSResolver}
			}
		}

		httpMap[keyStr] = entity.HTTPConfig{
			Front: entity.HTTPFront{
				Port:         item.Port,
				Hostname:     item.Hostname,
				HTTP:         item.HTTP,
				TLS:          item.TLS,
				H2:           item.H2,
				HSTS:         item.HSTS,
				Certificate:  item.Certificate,
				ProxyHeaders: proxyHeaders,
			},
			Feature: entity.HTTPFeature{
				Compress: item.Compress,
			},
			Rule: ruleConfigs,
			Backend: entity.HTTPBackend{
				RealIp:      item.RealIp,
				RootCA:      buildRootCAConfig(item.RootCA),
				Tunnel:      backendTunnel,
				DNSResolver: dnsResolver,
				Location:    locations,
			},
		}
	}

	// Build dedicated WebVPN Gateway services
	var webvpnServices []models.WebvpnService
	_ = engine.Where("status = 1").Find(&webvpnServices)
	for _, svc := range webvpnServices {
		svcKeyStr := "webvpn_" + strconv.FormatInt(svc.Id, 10)
		rootDomain := strings.TrimPrefix(svc.Hostname, "*.")

		var vpnSites []models.WebvpnSite
		_ = engine.Where("(service_id = ? OR (service_id = 0 AND http_proxy_id = ?)) AND status = 1", svc.Id, svc.Id).Find(&vpnSites)

		vpnActionSites := make(map[string]interface{})
		for _, vs := range vpnSites {
			u, err := url.Parse(vs.TargetURL)
			if err != nil {
				continue
			}
			targetHost := u.Hostname()
			targetPort := u.Port()
			schemePrefix := "s-"
			if u.Scheme == "http" {
				schemePrefix = "c-"
				if targetPort == "" {
					targetPort = "80"
				}
			} else {
				if targetPort == "" {
					targetPort = "443"
				}
			}

			hostMap := make(map[string]string)
			wildcardMap := make(map[string]string)

			// 1. Primary target host
			dashedTarget := strings.ReplaceAll(strings.ReplaceAll(targetHost, "-", "--"), ".", "-")
			mainVpnHost := fmt.Sprintf("%s%s-%s.%s", schemePrefix, dashedTarget, targetPort, rootDomain)
			hostMap[u.Host] = mainVpnHost
			if u.Port() != "" {
				hostMap[targetHost] = mainVpnHost
			}

			// 2. Related domain names from vs.Hosts
			for _, line := range strings.Split(vs.Hosts, "\n") {
				domain := strings.TrimSpace(line)
				if domain == "" {
					continue
				}
				if strings.Contains(domain, "*") {
					wildcardMap[domain] = ""
				} else {
					dashed := strings.ReplaceAll(strings.ReplaceAll(domain, "-", "--"), ".", "-")
					relVpnHost := fmt.Sprintf("%s%s-%s.%s", schemePrefix, dashed, targetPort, rootDomain)
					hostMap[domain] = relVpnHost
				}
			}

			isProt := vs.IsProtected == 1
			var groupIds []int64
			if isProt && vs.AllowedGroupIds != "" {
				_ = json.Unmarshal([]byte(vs.AllowedGroupIds), &groupIds)
			}

			replaceMap := make(map[string]string)
			if vs.Replace != "" {
				_ = json.Unmarshal([]byte(vs.Replace), &replaceMap)
			}

			vpnActionSites[vs.Prefix] = map[string]interface{}{
				"name":              vs.Name,
				"protected":         isProt,
				"allowed_group_ids": groupIds,
				"host":              hostMap,
				"wildcard":          wildcardMap,
				"replace":           replaceMap,
			}
		}

		loginURL := svc.LoginURL
		if loginURL == "" {
			loginURL = discoverAuthLoginURL(httpList, rulesMap)
		}

		fallbackPolicy := svc.Fallback
		if fallbackPolicy == "" {
			fallbackPolicy = "404"
		}

		ruleConfigs := []entity.RuleConfig{
			{
				Matcher: entity.MatcherConfig{
					Name:   "always_true_matcher",
					Config: map[string]interface{}{},
				},
				Action: entity.ActionConfig{
					Name: "subdomain_webvpn_action",
					Config: map[string]interface{}{
						"Sites":        vpnActionSites,
						"LoginURL":     loginURL,
						"CookieDomain": "." + rootDomain,
						"Fallback":     fallbackPolicy,
					},
				},
			},
		}

		httpMap[svcKeyStr] = entity.HTTPConfig{
			Front: entity.HTTPFront{
				Port:         svc.Port,
				Hostname:     svc.Hostname,
				HTTP:         true,
				TLS:          svc.TLS,
				H2:           svc.H2,
				Certificate:  svc.Certificate,
				ProxyHeaders: []string{},
			},
			Feature: entity.HTTPFeature{
				Compress: false,
			},
			Rule: ruleConfigs,
			Backend: entity.HTTPBackend{
				Location: []entity.HTTPLocation{
					{
						Path: "/",
						Upstream: entity.LocationUpstream{
							Type: "proxy_pass",
							Data: map[string]interface{}{
								"Method": "round_robin",
								"Servers": []map[string]interface{}{
									{"Target": "http://127.0.0.1:80", "Weight": 1},
								},
							},
						},
					},
				},
			},
		}
	}

	return httpMap
}

func discoverAuthLoginURL(httpList []models.HttpProxy, rulesMap map[string]models.Rule) string {
	for _, p := range httpList {
		if p.Rules != "" {
			var rNames []string
			_ = json.Unmarshal([]byte(p.Rules), &rNames)
			for _, rn := range rNames {
				if r, ok := rulesMap[rn]; ok && strings.Contains(r.Items, "auth_portal_action") {
					scheme := "http://"
					if p.TLS || p.H2 {
						scheme = "https://"
					}
					portPart := ""
					if p.Port != "80" && p.Port != "443" && p.Port != "" {
						portPart = ":" + p.Port
					}
					return scheme + p.Hostname + portPart
				}
			}
		}
	}

	for _, r := range rulesMap {
		if strings.Contains(r.Items, "auth_guard_action") {
			var items []entity.RuleConfig
			if err := json.Unmarshal([]byte(r.Items), &items); err == nil {
				for _, it := range items {
					if it.Action.Name == "auth_guard_action" {
						if cfgMap, ok := it.Action.Config.(map[string]interface{}); ok {
							if pu, ok := cfgMap["portal_url"].(string); ok && pu != "" {
								return pu
							}
						}
					}
				}
			}
		}
	}
	return ""
}

func buildTunnelClientMap() map[string]string {
	engine := models.GetEngine()
	if engine == nil {
		return nil
	}
	var clients []models.TunnelClient
	err := engine.Find(&clients)
	if err != nil {
		log.Printf("buildTunnelClientMap error: %v\n", err)
		return nil
	}

	clientMap := make(map[string]string)
	for _, c := range clients {
		token := strings.TrimSpace(c.Token)
		name := strings.TrimSpace(c.Name)
		if token == "" {
			continue
		}
		clientMap[token] = name
	}
	return clientMap
}

// BuildFullTunnelConfig builds the combined tunnel.json data structure matching ang engine
func BuildFullTunnelConfig() entity.TunnelFileConfig {
	tlsMap, quicMap := buildTunnelMaps()
	clientMap := buildTunnelClientMap()

	return entity.TunnelFileConfig{
		TLSTunnel:    tlsMap,
		QUICTunnel:   quicMap,
		TunnelClient: clientMap,
	}
}

// BuildFullCertificateConfig builds the certificate.json data structure
func BuildFullCertificateConfig() entity.CertificateFileConfig {
	certMap := buildCertMap()
	return entity.CertificateFileConfig{
		Certificate: certMap,
	}
}

func buildTCPMap(rulesMap map[string]models.Rule) map[string]entity.TCPConfig {
	engine := models.GetEngine()
	if engine == nil {
		return nil
	}
	var tcpList []models.TcpProxy
	err := engine.Find(&tcpList)
	if err != nil {
		log.Printf("buildTCPMap error: %v\n", err)
		return nil
	}

	tcpMap := make(map[string]entity.TCPConfig)
	for _, item := range tcpList {
		keyStr := strconv.FormatInt(item.Id, 10)

		// Parse Rules (Rule Set expansion)
		var ruleConfigs []entity.RuleConfig
		if item.Rules != "" {
			var ruleNames []string
			_ = json.Unmarshal([]byte(item.Rules), &ruleNames)
			for _, rName := range ruleNames {
				rName = strings.TrimSpace(rName)
				if dbRule, exists := rulesMap[rName]; exists {
					if dbRule.Items != "" {
						var items []entity.RuleConfig
						if err := json.Unmarshal([]byte(dbRule.Items), &items); err == nil {
							ruleConfigs = append(ruleConfigs, items...)
						}
					}
				}
			}
		}

		// Parse Backend
		var backend *entity.TCPBackend
		hasTunnel := item.TunnelId != ""
		var upstreamServers []entity.UpstreamServer
		if item.UpstreamServers != "" {
			_ = json.Unmarshal([]byte(item.UpstreamServers), &upstreamServers)
		}
		hasUpstream := len(upstreamServers) > 0

		if hasTunnel || hasUpstream {
			backend = &entity.TCPBackend{}
			if hasTunnel {
				backend.Tunnel = &entity.BackendTunnel{
					Type:  item.TunnelType,
					ID:    item.TunnelId,
					Token: item.TunnelToken,
				}
			}
			if hasUpstream {
				method := item.UpstreamMethod
				if method == "" {
					method = "round_robin"
				}
				backend.Upstream = &entity.UpstreamConfig{
					Method: method,
					Data: entity.UpstreamData{
						Servers: upstreamServers,
					},
				}
			}
		}

		tcpMap[keyStr] = entity.TCPConfig{
			Address: item.Address,
			Port:    item.Port,
			Rule:    ruleConfigs,
			Backend: backend,
		}
	}
	return tcpMap
}

// BuildFullServerConfig builds the combined server.json data structure matching ang engine
func BuildFullServerConfig() entity.ServerConfig {
	rulesMap := buildRulesDBMap()
	dnsMap := buildDNSMap(rulesMap)
	httpMap := buildHTTPMap(rulesMap)
	tcpMap := buildTCPMap(rulesMap)
	udpMap := buildUDPMap(rulesMap)
	sniMap := buildSNIMap(rulesMap)

	return entity.ServerConfig{
		DNS:  dnsMap,
		HTTP: httpMap,
		TCP:  tcpMap,
		UDP:  udpMap,
		SNI:  sniMap,
	}
}

// BuildFullGroupConfig builds the full GroupFileConfig from DB
func BuildFullGroupConfig() entity.GroupFileConfig {
	res := make(entity.GroupFileConfig)
	engine := models.GetEngine()
	if engine == nil {
		return res
	}
	var groups []models.UserGroup
	if err := engine.Find(&groups); err != nil {
		log.Printf("BuildFullGroupConfig error: %v\n", err)
		return res
	}
	for _, g := range groups {
		res[strconv.FormatInt(g.Id, 10)] = entity.GroupConfig{
			Id:          g.Id,
			Name:        g.Name,
			Description: g.Description,
			IsDefault:   g.IsDefault,
		}
	}
	return res
}

// BuildFullUserConfig builds the full UserFileConfig from DB
func BuildFullUserConfig() entity.UserFileConfig {
	res := make(entity.UserFileConfig)
	engine := models.GetEngine()
	if engine == nil {
		return res
	}
	var users []models.User
	if err := engine.Find(&users); err != nil {
		log.Printf("BuildFullUserConfig error: %v\n", err)
		return res
	}
	for _, u := range users {
		var gids []int64
		if u.GroupIds != "" {
			_ = json.Unmarshal([]byte(u.GroupIds), &gids)
		}
		if gids == nil {
			gids = []int64{}
		}
		res[u.Username] = entity.UserConfig{
			Id:         u.Id,
			Username:   u.Username,
			Password:   u.Password,
			FullName:   u.FullName,
			Email:      u.Email,
			Mobile:     u.Mobile,
			SourceType: u.SourceType,
			SourceId:   u.SourceId,
			GroupIds:   gids,
			Status:     u.Status,
			ExpireAt:   u.ExpireAt,
		}
	}
	return res
}

// PushPartialConfigToNodes pushes only the changed sections to all ang engine nodes.
// serverSections: map of top-level server.json keys to include (e.g. {"TCP": tcpMap}).
// Pass nil to skip pushing server_config.
//
// tunnelCfg: pass non-nil to include tunnel_config in the push.
// certCfg:   pass non-nil to include certificate_config in the push.
// userCfg:   pass non-nil to include user_config in the push.
// groupCfg:  pass non-nil to include group_config in the push.
//
// The ang engine's /api/config/sync endpoint applies a merge-patch — only the keys
// present in the payload are updated; everything else is left untouched on the node.
func PushPartialConfigToNodes(
	serverSections map[string]interface{},
	tunnelCfg *entity.TunnelFileConfig,
	certCfg *entity.CertificateFileConfig,
	userCfg *entity.UserFileConfig,
	groupCfg *entity.GroupFileConfig,
) {
	engine := models.GetEngine()
	if engine == nil {
		return
	}
	var nodes []models.ClusterNode
	if err := engine.Find(&nodes); err != nil || len(nodes) == 0 {
		return
	}

	payload := make(map[string]interface{})
	if len(serverSections) > 0 {
		payload["server_config"] = serverSections
	}
	if tunnelCfg != nil {
		payload["tunnel_config"] = tunnelCfg
	}
	if certCfg != nil {
		payload["certificate_config"] = certCfg
	}
	if userCfg != nil {
		payload["user_config"] = userCfg
	}
	if groupCfg != nil {
		payload["group_config"] = groupCfg
	}
	if len(payload) == 0 {
		return
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	client := &http.Client{Timeout: 5 * time.Second}
	for _, node := range nodes {
		addr := strings.TrimRight(node.Addr, "/")
		if addr == "" {
			continue
		}
		syncURL := addr + "/api/config/sync"
		req, err := http.NewRequest("POST", syncURL, bytes.NewBuffer(data))
		if err != nil {
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		if node.Secret != "" {
			req.Header.Set("X-Ang-Secret", node.Secret)
		}
		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			node.Status = 1
			node.LastPing = time.Now()
			_, _ = engine.ID(node.Id).Cols("status", "last_ping").Update(&node)
			log.Printf("[cluster_sync] Partial push to node %s (%s) OK", node.Name, node.Addr)
		} else {
			if resp != nil {
				resp.Body.Close()
			}
			node.Status = 0
			_, _ = engine.ID(node.Id).Cols("status").Update(&node)
			log.Printf("[cluster_sync] Partial push to node %s (%s) FAILED", node.Name, node.Addr)
		}
	}
}

// PushServerConfigToNodes pushes the full server.json, tunnel.json, certificate.json,
// user.json, and group.json to all registered ang engine nodes. Used by SyncAllToCluster.
func PushServerConfigToNodes(
	serverCfg entity.ServerConfig,
	tunnelCfg entity.TunnelFileConfig,
	certCfg entity.CertificateFileConfig,
	userCfg entity.UserFileConfig,
	groupCfg entity.GroupFileConfig,
) {
	engine := models.GetEngine()
	if engine == nil {
		return
	}
	var nodes []models.ClusterNode
	if err := engine.Find(&nodes); err != nil || len(nodes) == 0 {
		return
	}

	payload := map[string]interface{}{
		"server_config":      serverCfg,
		"tunnel_config":      tunnelCfg,
		"certificate_config": certCfg,
		"user_config":        userCfg,
		"group_config":       groupCfg,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, node := range nodes {
		addr := strings.TrimRight(node.Addr, "/")
		if addr == "" {
			continue
		}
		syncURL := addr + "/api/config/sync"
		req, err := http.NewRequest("POST", syncURL, bytes.NewBuffer(data))
		if err != nil {
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		if node.Secret != "" {
			req.Header.Set("X-Ang-Secret", node.Secret)
		}

		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			node.Status = 1
			node.LastPing = time.Now()
			_, _ = engine.ID(node.Id).Cols("status", "last_ping").Update(&node)
			log.Printf("[cluster_sync] Successfully pushed config to node %s (%s)", node.Name, node.Addr)
		} else {
			if resp != nil {
				resp.Body.Close()
			}
			node.Status = 0
			_, _ = engine.ID(node.Id).Cols("status").Update(&node)
			log.Printf("[cluster_sync] Failed to push config to node %s (%s)", node.Name, node.Addr)
		}
	}
}

// PingNode sends a heartbeat check to an ang engine node
func PingNode(node *models.ClusterNode) (bool, string) {
	return VerifyNode(node.Addr, node.Secret)
}

// SyncCertificateToCluster syncs certificate changes. Only certificate_config is pushed.
func SyncCertificateToCluster() {
	certCfg := BuildFullCertificateConfig()
	cluster.Put("Certificate", certCfg)
	cluster.PrintFullCertificateConfig(certCfg)
	go PushPartialConfigToNodes(nil, nil, &certCfg, nil, nil)
}

// SyncTunnelToCluster syncs tunnel changes. Only tunnel_config is pushed.
func SyncTunnelToCluster() {
	tunnelCfg := BuildFullTunnelConfig()
	cluster.Put("TUNNEL", tunnelCfg)
	cluster.PrintFullTunnelConfig(tunnelCfg)
	go PushPartialConfigToNodes(nil, &tunnelCfg, nil, nil, nil)
}

// SyncDNSToCluster syncs DNS proxy changes. Only the DNS section of server_config is pushed.
func SyncDNSToCluster() {
	rulesMap := buildRulesDBMap()
	dnsMap := buildDNSMap(rulesMap)
	cluster.Put("DNS", dnsMap)
	cluster.PrintFullServerConfig(BuildFullServerConfig())
	go PushPartialConfigToNodes(
		map[string]interface{}{"DNS": dnsMap},
		nil, nil, nil, nil,
	)
}

// SyncRuleToCluster syncs rule changes. Rules affect all server sections,
// so HTTP + TCP + DNS are all pushed together (but tunnel/cert unchanged).
func SyncRuleToCluster() {
	rulesMap := buildRulesDBMap()
	cluster.Put("Rule", rulesMap)
	serverCfg := BuildFullServerConfig()
	cluster.PrintFullServerConfig(serverCfg)
	go PushPartialConfigToNodes(
		map[string]interface{}{
			"HTTP": serverCfg.HTTP,
			"TCP":  serverCfg.TCP,
			"DNS":  serverCfg.DNS,
		},
		nil, nil, nil, nil,
	)
}

// SyncHTTPToCluster syncs HTTP proxy changes. Only the HTTP section of server_config is pushed.
func SyncHTTPToCluster() {
	rulesMap := buildRulesDBMap()
	httpMap := buildHTTPMap(rulesMap)
	cluster.Put("HTTP", httpMap)
	cluster.PrintFullServerConfig(BuildFullServerConfig())
	go PushPartialConfigToNodes(
		map[string]interface{}{"HTTP": httpMap},
		nil, nil, nil, nil,
	)
}

// SyncTCPToCluster syncs TCP proxy changes. Only the TCP section of server_config is pushed.
func SyncTCPToCluster() {
	rulesMap := buildRulesDBMap()
	tcpMap := buildTCPMap(rulesMap)
	cluster.Put("TCP", tcpMap)
	cluster.PrintFullServerConfig(BuildFullServerConfig())
	go PushPartialConfigToNodes(
		map[string]interface{}{"TCP": tcpMap},
		nil, nil, nil, nil,
	)
}

// SyncUserToCluster syncs user changes. Only user_config is pushed.
func SyncUserToCluster() {
	userCfg := BuildFullUserConfig()
	cluster.Put("USER", userCfg)
	cluster.PrintFullUserConfig(userCfg)
	go PushPartialConfigToNodes(nil, nil, nil, &userCfg, nil)
}

// SyncGroupToCluster syncs user group changes. Only group_config is pushed.
func SyncGroupToCluster() {
	groupCfg := BuildFullGroupConfig()
	cluster.Put("GROUP", groupCfg)
	cluster.PrintFullGroupConfig(groupCfg)
	go PushPartialConfigToNodes(nil, nil, nil, nil, &groupCfg)
}

// SyncUserAndGroupToCluster syncs both users and user groups.
func SyncUserAndGroupToCluster() {
	userCfg := BuildFullUserConfig()
	groupCfg := BuildFullGroupConfig()
	cluster.Put("USER", userCfg)
	cluster.Put("GROUP", groupCfg)
	cluster.PrintFullUserConfig(userCfg)
	cluster.PrintFullGroupConfig(groupCfg)
	go PushPartialConfigToNodes(nil, nil, nil, &userCfg, &groupCfg)
}

// SyncAllToCluster syncs everything to the cluster. Sends the full config payload.
func SyncAllToCluster() {
	certCfg := BuildFullCertificateConfig()
	tunnelCfg := BuildFullTunnelConfig()
	serverCfg := BuildFullServerConfig()
	userCfg := BuildFullUserConfig()
	groupCfg := BuildFullGroupConfig()

	cluster.Put("Certificate", certCfg)
	cluster.Put("TUNNEL", tunnelCfg)
	cluster.Put("USER", userCfg)
	cluster.Put("GROUP", groupCfg)
	cluster.PrintFullCertificateConfig(certCfg)
	cluster.PrintFullTunnelConfig(tunnelCfg)
	cluster.PrintFullServerConfig(serverCfg)
	cluster.PrintFullUserConfig(userCfg)
	cluster.PrintFullGroupConfig(groupCfg)

	go PushServerConfigToNodes(serverCfg, tunnelCfg, certCfg, userCfg, groupCfg)
}

func VerifyNode(addr, secret string) (bool, string) {
	addr = strings.TrimRight(addr, "/")
	if addr == "" {
		return false, "empty address"
	}
	verifyURL := addr + "/api/verify"
	client := &http.Client{Timeout: 3 * time.Second}
	req, err := http.NewRequest("GET", verifyURL, nil)
	if err != nil {
		return false, err.Error()
	}
	if secret != "" {
		req.Header.Set("X-Ang-Secret", secret)
	}
	resp, err := client.Do(req)
	if err != nil {
		return false, err.Error()
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		return true, "success"
	}
	if resp.StatusCode == 401 {
		return false, "auth_failed"
	}
	return false, fmt.Sprintf("http_status_%d", resp.StatusCode)
}

func buildUDPMap(rulesMap map[string]models.Rule) map[string]entity.UDPConfig {
	engine := models.GetEngine()
	if engine == nil {
		return nil
	}
	var udpList []models.UdpProxy
	err := engine.Find(&udpList)
	if err != nil {
		log.Printf("buildUDPMap error: %v\n", err)
		return nil
	}

	udpMap := make(map[string]entity.UDPConfig)
	for _, item := range udpList {
		keyStr := strconv.FormatInt(item.Id, 10)

		// Parse Rules (Rule Set expansion)
		var ruleConfigs []entity.RuleConfig
		if item.Rules != "" {
			var ruleNames []string
			_ = json.Unmarshal([]byte(item.Rules), &ruleNames)
			for _, rName := range ruleNames {
				rName = strings.TrimSpace(rName)
				if dbRule, exists := rulesMap[rName]; exists {
					if dbRule.Items != "" {
						var items []entity.RuleConfig
						if err := json.Unmarshal([]byte(dbRule.Items), &items); err == nil {
							ruleConfigs = append(ruleConfigs, items...)
						}
					}
				}
			}
		}

		// Parse Backend
		var backend *entity.UDPBackend
		hasTunnel := item.TunnelId != ""
		var upstreamServers []entity.UpstreamServer
		if item.UpstreamServers != "" {
			_ = json.Unmarshal([]byte(item.UpstreamServers), &upstreamServers)
		}
		hasUpstream := len(upstreamServers) > 0

		if hasTunnel || hasUpstream {
			backend = &entity.UDPBackend{}
			if hasTunnel {
				backend.Tunnel = &entity.BackendTunnel{
					Type:  item.TunnelType,
					ID:    item.TunnelId,
					Token: item.TunnelToken,
				}
			}
			if hasUpstream {
				method := item.UpstreamMethod
				if method == "" {
					method = "round_robin"
				}
				backend.Upstream = &entity.UpstreamConfig{
					Method: method,
					Data: entity.UpstreamData{
						Servers: upstreamServers,
					},
				}
			}
		}

		udpMap[keyStr] = entity.UDPConfig{
			Address: item.Address,
			Port:    item.Port,
			Rule:    ruleConfigs,
			Backend: backend,
		}
	}
	return udpMap
}

// SyncUDPToCluster syncs UDP proxy changes. Only the UDP section of server_config is pushed.
func SyncUDPToCluster() {
	rulesMap := buildRulesDBMap()
	udpMap := buildUDPMap(rulesMap)
	cluster.Put("UDP", udpMap)
	cluster.PrintFullServerConfig(BuildFullServerConfig())
	go PushPartialConfigToNodes(
		map[string]interface{}{"UDP": udpMap},
		nil, nil, nil, nil,
	)
}

func buildSNIMap(rulesMap map[string]models.Rule) map[string]entity.SNIConfig {
	engine := models.GetEngine()
	if engine == nil {
		return nil
	}
	var sniList []models.SniProxy
	err := engine.Find(&sniList)
	if err != nil {
		log.Printf("buildSNIMap error: %v\n", err)
		return nil
	}

	sniMap := make(map[string]entity.SNIConfig)
	for _, item := range sniList {
		keyStr := strconv.FormatInt(item.Id, 10)

		// Parse Rules (Rule Set expansion)
		var ruleConfigs []entity.RuleConfig
		if item.Rules != "" {
			var ruleNames []string
			_ = json.Unmarshal([]byte(item.Rules), &ruleNames)
			for _, rName := range ruleNames {
				rName = strings.TrimSpace(rName)
				if dbRule, exists := rulesMap[rName]; exists {
					if dbRule.Items != "" {
						var items []entity.RuleConfig
						if err := json.Unmarshal([]byte(dbRule.Items), &items); err == nil {
							ruleConfigs = append(ruleConfigs, items...)
						}
					}
				}
			}
		}

		// Parse Backend
		var backend *entity.SNIBackend
		hasTunnel := item.TunnelId != ""
		hasDnsResolver := item.DNSResolver != ""

		if hasTunnel || hasDnsResolver {
			backend = &entity.SNIBackend{}
			if hasTunnel {
				backend.Tunnel = &entity.BackendTunnel{
					Type:  item.TunnelType,
					ID:    item.TunnelId,
					Token: item.TunnelToken,
				}
			}
			if hasDnsResolver {
				var dnsResolvers []string
				if err := json.Unmarshal([]byte(item.DNSResolver), &dnsResolvers); err == nil {
					backend.DNSResolver = dnsResolvers
				} else {
					// Fallback for comma separated or single string
					parts := strings.Split(item.DNSResolver, ",")
					var cleaned []string
					for _, p := range parts {
						p = strings.TrimSpace(p)
						if p != "" {
							cleaned = append(cleaned, p)
						}
					}
					backend.DNSResolver = cleaned
				}
			}
		}

		// Parse ExtraSNI patterns
		var extraSNI []string
		if item.ExtraSNI != "" {
			_ = json.Unmarshal([]byte(item.ExtraSNI), &extraSNI)
		}

		sniMap[keyStr] = entity.SNIConfig{
			SNI:      item.SNI,
			ExtraSNI: extraSNI,
			Port:     item.Port,
			Rule:     ruleConfigs,
			Backend:  backend,
		}
	}
	return sniMap
}

// SyncSNIToCluster syncs SNI proxy changes. Only the SNI section of server_config is pushed.
func SyncSNIToCluster() {
	rulesMap := buildRulesDBMap()
	sniMap := buildSNIMap(rulesMap)
	cluster.Put("SNI", sniMap)
	cluster.PrintFullServerConfig(BuildFullServerConfig())
	go PushPartialConfigToNodes(
		map[string]interface{}{"SNI": sniMap},
		nil, nil, nil, nil,
	)
}
