package vars

import "sync"

var (
	v     map[string]string
	vOnce sync.Once
)

func Setup() {
	vOnce.Do(func() {
		v = map[string]string{
			"BRAND_NAME":             "fast vpn",
			"TG_CHANNEL":             "@FastVPN",
			"MAIN_KB_INLINE":         "false",
			"TELEGRAM_BOT_TOKEN":     "7451072350:AAHWsuBPz28dkWxNYBo4YRKw_nRFnJq4yWE",
			"MYSQL_USER":             "root",
			"MYSQL_PASS":             "root",
			"MYSQL_HOST":             "localhost",
			"MYSQL_PORT":             "3306",
			"MYSQL_DB":               "vpn",
			"STORAGE_CHANNEL_ID":     "@techybattest",
			"DEFAULT_ADMIN_USERNAME": "sinatechs",
			"PANEL_URL":              "https://example.com/",
			"PANEL_SUB_URL":          "example.com",
			"PANEL_SUB_PORT":         "8443",
			"PANEL_SUB_PATH":         "/sub/",
			"PANEL_JSON_SUB_PATH":    "/json/",
			"PANEL_USERNAME":         "admin",
			"PANEL_PASSWORD":         "admin",
			"CONFIG_HOST":            "example.com",
			"ONLY_TRUSTED_USERS":     "false",

			"SUB_URL":           "https://example.com/",
			"SUB_PORT":          "2087",
			"SUB_PATH":          "/sub/",
			"SUB_CERT_FILE":     "/root/cert/origin/pub.cert",
			"SUB_CERT_KEY_FILE": "/root/cert/origin/priv.pem",

			"WALLET_INIT_BALANCE": "10000",
			"LOGO_PATH":           "vpn.png",
			"QR_PATH":             "qr",
		}
	})
}

func Get(name string) string {
	Setup()
	return v[name]
}
