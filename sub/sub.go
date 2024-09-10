package sub

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm/clause"
	"techybat.org/go-vpn/database"
	m "techybat.org/go-vpn/models"
	"techybat.org/go-vpn/panel"
)

type Vless struct {
	title, ip, operator, subID, shortLink string
}

type Config interface {
	stringify()
}

func (vless Vless) stringify() string {
	temp := strings.Split(vless.shortLink, "://")
	protocol, temp := temp[0], strings.Split(temp[1], "@")
	uuid, temp := temp[0], strings.Split(temp[1], ":")
	rest := strings.Split(temp[1], "#")[0]
	configDescriptor := fmt.Sprintf("%s://%s@%s:%s#%s", protocol, uuid, vless.ip, rest, vless.title)
	return configDescriptor
}

func handler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		http.Error(w, "sub id is required", http.StatusBadRequest)
		fmt.Printf("requested Subscription failed: subid:%s\n", id)
		return
	}
	title, _ := getTitle(id)
	short, _ := panel.ShortLinkConfigFromSubID(title, id)
	vlessStrings, _ := getStatusConfigs(id)
	for _, config := range extractIPs(readLines()) {
		config.title = title + " | " + config.operator
		config.subID = id
		config.shortLink = short
		vlessStrings = append(vlessStrings, config.stringify())
	}

	response := strings.Join(vlessStrings, "\n")
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(response))
	fmt.Printf("Requested Subscription: subID:%s\n", id)
}

func getStatusConfigs(subID string) ([]string, error) {
	db := database.GetDB()
	statusConfigs := []string{}
	var config m.Config
	if res := db.Preload(clause.Associations).Preload("Order.Pack").Find(&config, "sub_id = ?", subID); res.Error != nil {
		return statusConfigs, res.Error
	}

	traffic, err := config.TrafficName(db)
	if err != nil {
		return statusConfigs, err
	}
	trafficTitle := fmt.Sprintf("📊 %s / %s | #%d", traffic, config.Order.Pack.TrafficName(), config.OrderID)
	trafficTitle = url.QueryEscape(trafficTitle)
	statusConfigs = append(statusConfigs, fmt.Sprintf("trojan://uuid@1.1.1.1:2020?security=tls&headerType=none&type=tcp#%s", trafficTitle))
	endDate, err := config.EndDate(db)
	if err != nil {
		return statusConfigs, err
	}
	remainedTime, _ := config.RemainedDateStr(db)
	statusConfigs = append(statusConfigs, fmt.Sprintf("🗓 %s | %s | #%d", endDate.Format("y/MM/dd"), remainedTime, config.OrderID))
	statusConfigs[1] = fmt.Sprintf("trojan://uuid@1.1.1.1:2020?security=tls&headerType=none&type=tcp#%s", url.QueryEscape(statusConfigs[1]))
	return statusConfigs, nil
}

func getTitle(subID string) (string, error) {
	db := database.GetDB()
	var config m.Config
	if res := db.Find(&config, "sub_id = ?", subID); res.Error != nil {
		return "title not loaded", res.Error
	}

	return config.Title(db), nil
}

func readLines() []string {
	byteIps, err := os.ReadFile("clean_ips.txt")
	if err != nil {
		panic(err)
	}
	ipsStr := string(byteIps)
	return strings.Split(ipsStr, "\n")
}

func extractIPs(lines []string) []Vless {
	configs := []Vless{}
	for _, line := range lines {
		lineArray := strings.Split(line, " ")
		if len(lineArray) < 2 {
			continue
		}
		vless := Vless{}
		vless.ip = lineArray[0]
		vless.operator = lineArray[1]
		configs = append(configs, vless)
	}
	return configs
}

func ServeHttp(ctx context.Context) {
	path := os.Getenv("SUB_PATH")
	port := os.Getenv("SUB_PORT")
	r := mux.NewRouter()
	r.HandleFunc(fmt.Sprintf("/%s/{id}", path), handler)

	fmt.Printf("Sub Server is running on port %s...\n", port)

	srv := &http.Server{
		Addr: "0.0.0.0:" + port,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			fmt.Printf("Error starting server: %s\n", err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Sub server is going to shutdown...")
			srv.Shutdown(ctx)
			return
		}
	}
}

func ServeHttps(ctx context.Context) {
	path := os.Getenv("SUB_PATH")
	port := os.Getenv("SUB_PORT")
	r := mux.NewRouter()
	r.HandleFunc(fmt.Sprintf("/%s/{id}", path), handler)

	fmt.Printf("Sub Server is running on port %s with TLS...\n", port)

	// Path to your SSL certificate and key files
	certFile := os.Getenv("SUB_CERT_FILE")
	keyFile := os.Getenv("SUB_CERT_KEY_FILE")

	srv := &http.Server{
		Addr: "0.0.0.0:" + port,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}

	go func() {
		if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil {
			fmt.Printf("Error starting server: %s\n", err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Sub server is going to shutdown...")
			srv.Shutdown(ctx)
			return
		}
	}
}
