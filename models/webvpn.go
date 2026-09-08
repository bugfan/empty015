package models

import "time"

type WebvpnService struct {
	Id          int64     `xorm:"pk autoincr 'id'" json:"id"`
	Name        string    `xorm:"varchar(255) notnull 'name'" json:"name"`                                 // 服务名称，如 "校园主网关"
	Hostname    string    `xorm:"varchar(255) notnull 'hostname'" json:"hostname"`                         // 泛域名，如 "*.webvpn.example.com"
	Port        string    `xorm:"varchar(10) default '443' 'port'" json:"port"`                            // 监听端口，默认 443
	TLS         bool      `xorm:"bool default true 'tls'" json:"tls"`                                      // 是否启用 TLS
	H2          bool      `xorm:"bool default true 'h2'" json:"h2"`                                        // 是否启用 HTTP/2
	Certificate string    `xorm:"varchar(255) 'certificate'" json:"certificate"`                           // 关联证书
	LoginURL    string    `xorm:"varchar(512) 'login_url'" json:"login_url"`                               // 认证中心登录地址
	Fallback    string    `xorm:"varchar(32) default '404' 'fallback'" json:"fallback"`                    // 未匹配兜底策略: "404" 或 "login"
	Status      int       `xorm:"tinyint default 1 'status'" json:"status"`                                // 1: 启用, 0: 禁用
	Remark      string    `xorm:"varchar(255) 'remark'" json:"remark"`                                     // 备注
	CreatedAt   time.Time `xorm:"created" json:"created_at"`
	UpdatedAt   time.Time `xorm:"updated" json:"updated_at"`
}

func (WebvpnService) TableName() string {
	return "webvpn_service"
}

func (w *WebvpnService) BeforeInsert() {
	if w.CreatedAt.IsZero() {
		w.CreatedAt = time.Now()
	}
	if w.UpdatedAt.IsZero() {
		w.UpdatedAt = time.Now()
	}
}

func (w *WebvpnService) BeforeUpdate() {
	w.UpdatedAt = time.Now()
}

type WebvpnSite struct {
	Id              int64     `xorm:"pk autoincr 'id'" json:"id"`
	Name            string    `xorm:"varchar(255) notnull 'name'" json:"name"`                                 // 资源应用名称，如 "中国知网"
	ServiceId       int64     `xorm:"index 'service_id'" json:"service_id"`                                    // 关联的 WebVPN 服务 ID
	HttpProxyId     int64     `xorm:"index 'http_proxy_id'" json:"http_proxy_id"`                              // 兼容保留字段
	TargetURL       string    `xorm:"varchar(512) notnull 'target_url'" json:"target_url"`                     // 目标真实地址，如 https://www.cnki.net
	Prefix          string    `xorm:"varchar(128) notnull 'prefix'" json:"prefix"`                             // 子域名前缀，如 s-cnki 或 s-www-cnki-net-443
	Hosts           string    `xorm:"text 'hosts'" json:"hosts"`                                               // 关联地址/代理域名列表 (每行一个域名)
	Replace         string    `xorm:"text 'replace'" json:"replace"`                                           // 扩展内容替换 JSON map, 如 {"知网": "XX"}
	AllowedGroupIds string    `xorm:"varchar(255) default '[]' 'allowed_group_ids'" json:"allowed_group_ids"` // 允许访问的用户组 ID 列表 JSON 数组
	IsProtected     int       `xorm:"tinyint default 1 'is_protected'" json:"is_protected"`                    // 1: 保护模式(需登录), 0: 公开模式(免登录)
	Status          int       `xorm:"tinyint default 1 'status'" json:"status"`                                // 1: 启用, 0: 禁用
	Remark          string    `xorm:"varchar(255) 'remark'" json:"remark"`                                     // 备注
	CreatedAt       time.Time `xorm:"created" json:"created_at"`
	UpdatedAt       time.Time `xorm:"updated" json:"updated_at"`
}

func (WebvpnSite) TableName() string {
	return "webvpn_site"
}

func (w *WebvpnSite) BeforeInsert() {
	if w.CreatedAt.IsZero() {
		w.CreatedAt = time.Now()
	}
	if w.UpdatedAt.IsZero() {
		w.UpdatedAt = time.Now()
	}
}

func (w *WebvpnSite) BeforeUpdate() {
	w.UpdatedAt = time.Now()
}
