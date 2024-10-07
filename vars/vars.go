package vars

import "sync"

var (
	v     map[string]string
	vOnce sync.Once
)

func Setup() {
	vOnce.Do(func() {
		v = map[string]string{
			"BRAND_NAME":         "☄️ Ultra Fast VPN☄️",
			"TG_CHANNEL":         "@ultrafast_vpn",
			"MAIN_KB_INLINE":     "false",
			"TELEGRAM_BOT_TOKEN": "7462326358:AAHcdFLPV5d_ftm0c8XRtFdWSAfeR8Mkr3Q",
			"MYSQL_USER":         "root",
			"MYSQL_PASS":         "q634E&@PXOnu",
			"MYSQL_HOST":         "localhost",
			"MYSQL_PORT":         "3306",
			"MYSQL_DB":           "vpn_bot",

			"STORAGE_CHANNEL_ID":     "-1002387187163",
			"DEFAULT_ADMIN_USERNAME": "ultrafast24",

			"PANEL_URL":           "https://projectservic.ir:2053/admin/",
			"PANEL_SUB_URL":       "projectservic.ir",
			"PANEL_SUB_PORT":      "2096",
			"PANEL_SUB_PATH":      "/sub/",
			"PANEL_JSON_SUB_PATH": "/json/",
			"PANEL_USERNAME":      "ultrafast",
			"PANEL_PASSWORD":      "@uo5B2.I7i95",
			"CONFIG_HOST":         "projectservic.ir",
			"ONLY_TRUSTED_USERS":  "false",

			"SUB_URL":           "https://projectservic.ir:2053/admin/",
			"SUB_PORT":          "8443",
			"SUB_PATH":          "/sub/",
			"SUB_CERT_FILE":     "/root/cert/projectservic.ir/fullchain.pem",
			"SUB_CERT_KEY_FILE": "/root/cert/projectservic.ir/privkey.pem",

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
