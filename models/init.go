package models

import (
	"log"
	"time"

	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var engine *xorm.Engine

func InitDB(dsn string) {
	var err error
	engine, err = xorm.NewEngine("sqlite3", dsn)
	if err != nil {
		log.Fatalf("Failed to create engine: %v", err)
	}

	// Drop old tunnel table if it still contains obsolete key_name column
	if isExist, _ := engine.IsTableExist("tunnel"); isExist {
		results, err := engine.QueryString("PRAGMA table_info(tunnel)")
		if err == nil {
			for _, row := range results {
				if row["name"] == "key_name" {
					_ = engine.DropTables("tunnel")
					break
				}
			}
		}
	}

	// Migrate dns_provider table to acme_account if exists
	if isExist, _ := engine.IsTableExist("dns_provider"); isExist {
		_, _ = engine.Exec("ALTER TABLE dns_provider RENAME TO acme_account")
	}

	// Rename dns_provider_id to acme_account_id in certificate table
	if isExist, _ := engine.IsTableExist("certificate"); isExist {
		results, err := engine.QueryString("PRAGMA table_info(certificate)")
		if err == nil {
			for _, row := range results {
				if row["name"] == "dns_provider_id" {
					_, _ = engine.Exec("ALTER TABLE certificate RENAME COLUMN dns_provider_id TO acme_account_id")
					break
				}
			}
		}
	}

	// Migrate legacy tables if exists
	if isExist, _ := engine.IsTableExist("access_group"); isExist {
		_, _ = engine.Exec("ALTER TABLE access_group RENAME TO user_group")
	}
	if isExist, _ := engine.IsTableExist("access_user"); isExist {
		_, _ = engine.Exec("ALTER TABLE access_user RENAME TO user")
	}
	if isExist, _ := engine.IsTableExist("auth_source"); isExist {
		_, _ = engine.Exec("ALTER TABLE auth_source RENAME TO auth_method")
	}

	// Automatically sync database schemas if necessary
	err = engine.Sync2(
		new(AdminUser), new(Tunnel), new(Certificate), new(TunnelClient),
		new(DnsProxy), new(TcpProxy), new(UdpProxy), new(SniProxy),
		new(Rule), new(HttpProxy), new(ClusterNode), new(AcmeAccount),
		new(UserGroup), new(User), new(AuthMethod), new(Auth),
		new(WebvpnService), new(WebvpnSite),
	)

	if err != nil {
		log.Fatalf("Failed to sync database: %v", err)
	}

	// Migrate legacy WebvpnSite http_proxy_id to WebvpnService if webvpn_service is empty
	if count, err := engine.Count(new(WebvpnService)); err == nil && count == 0 {
		var sites []WebvpnSite
		_ = engine.Find(&sites)
		proxySvcMap := make(map[int64]int64)
		for _, s := range sites {
			if s.HttpProxyId > 0 && s.ServiceId == 0 {
				if svcId, ok := proxySvcMap[s.HttpProxyId]; ok {
					s.ServiceId = svcId
					_, _ = engine.ID(s.Id).Cols("service_id").Update(&s)
				} else {
					var p HttpProxy
					if has, _ := engine.ID(s.HttpProxyId).Get(&p); has {
						newSvc := WebvpnService{
							Name:        p.Name,
							Hostname:    p.Hostname,
							Port:        p.Port,
							TLS:         p.TLS,
							H2:          p.H2,
							Certificate: p.Certificate,
							Fallback:    "404",
							Status:      1,
							Remark:      p.Remark,
						}
						if _, err := engine.Insert(&newSvc); err == nil && newSvc.Id > 0 {
							proxySvcMap[s.HttpProxyId] = newSvc.Id
							s.ServiceId = newSvc.Id
							_, _ = engine.ID(s.Id).Cols("service_id").Update(&s)
						}
					}
				}
			}
		}
	}

	// Ensure default UserGroup exists
	if count, err := engine.Count(new(UserGroup)); err == nil && count == 0 {
		_, _ = engine.Insert(&UserGroup{
			Name:        "默认用户组",
			Description: "系统默认用户组",
			IsDefault:   true,
		})
	}

	// Ensure default Local AuthMethod exists
	if count, err := engine.Count(new(AuthMethod)); err == nil && count == 0 {
		_, _ = engine.Insert(&AuthMethod{
			Name:       "本地账号认证",
			Type:       "local",
			Enabled:    true,
			Priority:   1,
			ConfigJSON: `{"allow_self_register":false,"password_min_len":6}`,
			Remark:     "系统默认本地用户名密码认证",
		})
	}

	// Backfill certificate parsed metadata for all certificates
	var allCerts []Certificate
	if err := engine.Where("cert_content != ''").Find(&allCerts); err == nil {
		for _, cert := range allCerts {
			cert.ParseCertInfo()
			_, _ = engine.ID(cert.Id).Cols("subject_cn", "sans", "not_before", "not_after", "issuer", "serial_number").Update(&cert)
		}
	}

	// Backfill missing or zero created_at dates across tables
	nowStr := time.Now().Format("2006-01-02 15:04:05")
	_, _ = engine.Exec("UPDATE http_proxy SET created_at = ? WHERE created_at IS NULL OR created_at = '' OR created_at LIKE '0001-01-01%'", nowStr)
	_, _ = engine.Exec("UPDATE rule SET created_at = ? WHERE created_at IS NULL OR created_at = '' OR created_at LIKE '0001-01-01%'", nowStr)
	_, _ = engine.Exec("UPDATE dns_proxy SET created_at = ? WHERE created_at IS NULL OR created_at = '' OR created_at LIKE '0001-01-01%'", nowStr)
	_, _ = engine.Exec("UPDATE tunnel SET created_at = ? WHERE created_at IS NULL OR created_at = '' OR created_at LIKE '0001-01-01%'", nowStr)
	_, _ = engine.Exec("UPDATE certificate SET source = 'MANUAL' WHERE source IS NULL OR source = ''")
	_, _ = engine.Exec("UPDATE certificate SET source = 'MANUAL' WHERE source = 'SELF_SIGNED'")
	_, _ = engine.Exec("UPDATE certificate SET type = 'STD' WHERE type = 'SELF-STD'")
	_, _ = engine.Exec("UPDATE certificate SET source = 'ACME' WHERE cert_id LIKE 'acme-%'")

	// Initialize default admin user
	admin := &AdminUser{Username: "admin"}
	has, err := engine.Get(admin)
	if err != nil {
		log.Fatalf("Failed to query admin user: %v", err)
	}
	if !has {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin@123"), bcrypt.DefaultCost)
		admin.Password = string(hashedPassword)
		admin.IsSuperAdmin = true
		_, err = engine.Insert(admin)
		if err != nil {
			log.Fatalf("Failed to insert default admin: %v", err)
		}
	} else if !admin.IsSuperAdmin {
		admin.IsSuperAdmin = true
		_, _ = engine.ID(admin.Id).Cols("is_super_admin").Update(admin)
	}
}

func GetEngine() *xorm.Engine {
	return engine
}

