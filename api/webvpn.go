package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/bugfan/ang-admin/models"
	"github.com/bugfan/ang-admin/service"
	"github.com/bugfan/rest"
	"github.com/gin-gonic/gin"
	"github.com/go-xorm/xorm"
)

func init() {
	rest.Register(&models.WebvpnService{}, &webvpnServiceHandler{}, rest.RouteTypeALL, nil, "webvpn-service")
	rest.Register(&models.WebvpnSite{}, &webvpnSiteHandler{}, rest.RouteTypeALL, nil, "webvpn-site")
}

// -------------------------------------------------------------
// 1. WebVPN 服务 (WebvpnService) Handler
// -------------------------------------------------------------

type webvpnServiceHandler struct {
	Id          int64     `json:"id"`
	Name        string    `json:"name"`
	Hostname    string    `json:"hostname"`
	Port        string    `json:"port"`
	TLS         bool      `json:"tls"`
	H2          bool      `json:"h2"`
	Certificate string    `json:"certificate"`
	LoginURL    string    `json:"login_url"`
	Fallback    string    `json:"fallback"`
	Status      int       `json:"status"`
	Remark      string    `json:"remark"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (h *webvpnServiceHandler) Before(g *gin.Context, x *xorm.Engine) bool {
	method := g.Request.Method
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
		h.Name = strings.TrimSpace(h.Name)
		if h.Name == "" {
			g.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code":    1,
				"message": "服务名称不能为空",
			})
			return false
		}

		h.Hostname = strings.TrimSpace(h.Hostname)
		if h.Hostname == "" {
			g.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code":    1,
				"message": "泛域名不能为空",
			})
			return false
		}
		// 规范化通配符域名格式，如用户输入 example.com 或 webvpn.example.com，确保以 *. 开头
		if !strings.HasPrefix(h.Hostname, "*.") {
			if strings.HasPrefix(h.Hostname, "*") {
				h.Hostname = "*." + strings.TrimPrefix(h.Hostname, "*")
			} else {
				h.Hostname = "*." + h.Hostname
			}
		}

		h.Port = strings.TrimSpace(h.Port)
		if h.Port == "" {
			h.Port = "443"
		}

		h.Fallback = strings.TrimSpace(h.Fallback)
		if h.Fallback == "" {
			h.Fallback = "404"
		}
	}
	return true
}

func (h *webvpnServiceHandler) After(g *gin.Context, x *xorm.Engine, args ...interface{}) {
	method := g.Request.Method
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch || method == http.MethodDelete {
		service.SyncHTTPToCluster()
	}
}

func (h *webvpnServiceHandler) List(c *gin.Context) {
	var list []models.WebvpnService
	session := models.GetEngine().NewSession()
	defer session.Close()

	if name := strings.TrimSpace(c.Query("name")); name != "" {
		session.Where("name LIKE ?", "%"+name+"%")
	}
	if hostname := strings.TrimSpace(c.Query("hostname")); hostname != "" {
		session.Where("hostname LIKE ?", "%"+hostname+"%")
	}

	err := session.Desc("id").Find(&list)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}

	// 统计关联的站点数量
	type WebvpnServiceVO struct {
		models.WebvpnService
		RootDomain string `json:"root_domain"`
		SiteCount  int64  `json:"site_count"`
	}

	resList := make([]WebvpnServiceVO, len(list))
	for i, svc := range list {
		count, _ := models.GetEngine().Where("service_id = ?", svc.Id).Count(new(models.WebvpnSite))
		resList[i] = WebvpnServiceVO{
			WebvpnService: svc,
			RootDomain:    strings.TrimPrefix(svc.Hostname, "*."),
			SiteCount:     count,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    resList,
	})
}

// -------------------------------------------------------------
// 2. WebVPN 站点 (WebvpnSite) Handler
// -------------------------------------------------------------

type webvpnSiteHandler struct {
	Id              int64     `json:"id"`
	Name            string    `json:"name"`
	ServiceId       int64     `json:"service_id"`
	HttpProxyId     int64     `json:"http_proxy_id"`
	TargetURL       string    `json:"target_url"`
	Prefix          string    `json:"prefix"`
	Hosts           string    `json:"hosts"`
	Replace         string    `json:"replace"`
	AllowedGroupIds string    `json:"allowed_group_ids"`
	IsProtected     int       `json:"is_protected"`
	Status          int       `json:"status"`
	Remark          string    `json:"remark"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (h *webvpnSiteHandler) Before(g *gin.Context, x *xorm.Engine) bool {
	method := g.Request.Method
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
		h.Name = strings.TrimSpace(h.Name)
		if h.Name == "" {
			g.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code":    1,
				"message": "应用名称不能为空",
			})
			return false
		}

		// 兼容 service_id 与旧版 http_proxy_id
		if h.ServiceId <= 0 && h.HttpProxyId > 0 {
			h.ServiceId = h.HttpProxyId
		}
		if h.ServiceId <= 0 {
			g.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code":    1,
				"message": "必须选择所属的 WebVPN 服务",
			})
			return false
		}

		h.TargetURL = strings.TrimSpace(h.TargetURL)
		if h.TargetURL == "" || (!strings.HasPrefix(h.TargetURL, "http://") && !strings.HasPrefix(h.TargetURL, "https://")) {
			g.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code":    1,
				"message": "目标地址必须是以 http:// 或 https:// 开头的合法 URL",
			})
			return false
		}

		// 标准 WebVPN 子域名前缀推导: s-<dashed-host>-<port> (http为 c-)
		u, err := url.Parse(h.TargetURL)
		if err == nil {
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
			dashed := strings.ReplaceAll(strings.ReplaceAll(targetHost, "-", "--"), ".", "-")
			h.Prefix = fmt.Sprintf("%s%s-%s", schemePrefix, dashed, targetPort)
		}

		if h.AllowedGroupIds == "" {
			h.AllowedGroupIds = "[]"
		}
		if h.Replace == "" {
			h.Replace = "{}"
		}
	}
	return true
}

func (h *webvpnSiteHandler) After(g *gin.Context, x *xorm.Engine, args ...interface{}) {
	method := g.Request.Method
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch || method == http.MethodDelete {
		service.SyncHTTPToCluster()
	}
}

func (h *webvpnSiteHandler) List(c *gin.Context) {
	var list []models.WebvpnSite
	session := models.GetEngine().NewSession()
	defer session.Close()

	if name := strings.TrimSpace(c.Query("name")); name != "" {
		session.Where("name LIKE ?", "%"+name+"%")
	}
	if serviceIdStr := strings.TrimSpace(c.Query("service_id")); serviceIdStr != "" {
		if sid, err := strconv.ParseInt(serviceIdStr, 10, 64); err == nil && sid > 0 {
			session.Where("service_id = ?", sid)
		}
	} else if proxyIdStr := strings.TrimSpace(c.Query("http_proxy_id")); proxyIdStr != "" {
		if pid, err := strconv.ParseInt(proxyIdStr, 10, 64); err == nil && pid > 0 {
			session.Where("service_id = ? OR http_proxy_id = ?", pid, pid)
		}
	}

	err := session.Desc("id").Find(&list)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    1,
			"message": err.Error(),
		})
		return
	}

	// 预加载 WebvpnService 映射
	var services []models.WebvpnService
	_ = models.GetEngine().Find(&services)
	serviceMap := make(map[int64]models.WebvpnService)
	for _, s := range services {
		serviceMap[s.Id] = s
	}

	// 兼容旧版 HttpProxy 映射
	var proxies []models.HttpProxy
	_ = models.GetEngine().Find(&proxies)
	proxyMap := make(map[int64]models.HttpProxy)
	for _, p := range proxies {
		proxyMap[p.Id] = p
	}

	type WebvpnSiteItemVO struct {
		models.WebvpnSite
		ServiceName     string `json:"service_name"`
		ServiceHostname string `json:"service_hostname"`
		FullAccessURL   string `json:"full_access_url"`
	}

	resList := make([]WebvpnSiteItemVO, len(list))
	for i, item := range list {
		vo := WebvpnSiteItemVO{WebvpnSite: item}

		// 优先从 WebvpnService 关联
		if svc, ok := serviceMap[item.ServiceId]; ok {
			vo.ServiceName = svc.Name
			vo.ServiceHostname = svc.Hostname

			rootDomain := strings.TrimPrefix(svc.Hostname, "*.")
			scheme := "http://"
			if svc.TLS || svc.H2 {
				scheme = "https://"
			}
			portSuffix := ""
			if svc.Port != "80" && svc.Port != "443" && svc.Port != "" {
				portSuffix = ":" + svc.Port
			}
			vo.FullAccessURL = fmt.Sprintf("%s%s.%s%s", scheme, item.Prefix, rootDomain, portSuffix)
		} else if p, ok := proxyMap[item.HttpProxyId]; ok {
			// 兼容回退 HttpProxy
			vo.ServiceName = p.Name
			vo.ServiceHostname = p.Hostname

			rootDomain := strings.TrimPrefix(p.Hostname, "*.")
			scheme := "http://"
			if p.TLS || p.H2 {
				scheme = "https://"
			}
			portSuffix := ""
			if p.Port != "80" && p.Port != "443" && p.Port != "" {
				portSuffix = ":" + p.Port
			}
			vo.FullAccessURL = fmt.Sprintf("%s%s.%s%s", scheme, item.Prefix, rootDomain, portSuffix)
		}

		resList[i] = vo
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    resList,
	})
}

