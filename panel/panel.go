package panel

import (
	"github.com/go-resty/resty/v2"
	"techybat.org/go-vpn/vars"
)

type BasicResponse struct {
	Success bool   `json:"success"`
	Message string `json:"msg"`
}

type Panel struct {
	client *resty.Client
}

var (
	panel *Panel
)

func GetPanel() *Panel {
	if panel == nil {
		Setup()
	}

	return panel
}

func RevokePanel() {
	panel = nil
}

func Setup() {
	username, password := vars.Get("PANEL_USERNAME"), vars.Get("PANEL_PASSWORD")
	url := vars.Get("PANEL_URL")
	panel = &Panel{
		client: resty.New(),
	}

	panel.client.SetBaseURL(url)

	if err := panel.Login(username, password); err != nil {
		panic(err)
	}
}
