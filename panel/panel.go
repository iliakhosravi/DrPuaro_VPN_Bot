package panel

import (
	"os"
	"sync"

	"github.com/go-resty/resty/v2"
)

type BasicResponse struct {
	Success bool   `json:"success"`
	Message string `json:"msg"`
}

type Panel struct {
	client *resty.Client
}

var (
	panel     *Panel
	panelOnce sync.Once
)

func GetPanel() *Panel {
	panelOnce.Do(func() {
		Setup()
	})

	return panel
}

func Setup() {
	username, password := os.Getenv("PANEL_USERNAME"), os.Getenv("PANEL_PASSWORD")
	url := os.Getenv("PANEL_URL")
	panel = &Panel{
		client: resty.New(),
	}

	panel.client.SetBaseURL(url)

	if err := panel.Login(username, password); err != nil {
		panic(err)
	}
}
