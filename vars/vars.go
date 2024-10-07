package vars

import "sync"

var (
	v     map[string]string
	vOnce sync.Once
)

func Setup() {
	vOnce.Do(func() {
		v = map[string]string{
			"BRAND_NAME":         "fast vpn",
			"TG_CHANNEL":         "@FastVPN",
			"MAIN_KB_INLINE":     "false",
			"TELEGRAM_BOT_TOKEN": "7451072350:AAHWsuBPz28dkWxNYBo4YRKw_nRFnJq4yWE",
			"MYSQL_USER":         "root",
			"MYSQL_PASS":         "root",
			"MYSQL_HOST":         "localhost",
			"MYSQL_PORT":         "3306",
			"MYSQL_DB":           "vpn",

			"STORAGE_CHANNEL_ID":     "@techybattest",
			"DEFAULT_ADMIN_USERNAME": "sinatechs",

			"PANEL_URL":           "https://techybat.org:2083/xray-admin/",
			"PANEL_SUB_URL":       "sub.techybat.org",
			"PANEL_SUB_PORT":      "2096",
			"PANEL_SUB_PATH":      "/subs/",
			"PANEL_JSON_SUB_PATH": "/jsons/",
			"PANEL_USERNAME":      "sinasadeghi83",
			"PANEL_PASSWORD":      "5Si015730na",
			"CONFIG_HOST":         "techybat.org",
			"ONLY_TRUSTED_USERS":  "false",

			"SUB_URL":           "localhost",
			"SUB_PORT":          "2087",
			"SUB_PATH":          "/sub/",
			"SUB_CERT_FILE":     "/root/cert//fullchain.pem",
			"SUB_CERT_KEY_FILE": "/root/cert//privkey.pem",

			"WALLET_INIT_BALANCE": "10000",
			"LOGO_PATH":           "vpn.png",
			"QR_PATH":             "qr",
			"env":                 "prod",
		}
	})
}

func Get(name string) string {
	Setup()
	return v[name]
}
