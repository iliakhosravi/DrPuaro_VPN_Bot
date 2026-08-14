package panel

import (
	"sync"

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
	panel   *Panel
	panelMu sync.RWMutex
)

func GetPanel() *Panel {
	panelMu.RLock()
	current := panel
	panelMu.RUnlock()

	if current != nil {
		return current
	}

	Setup()

	panelMu.RLock()
	defer panelMu.RUnlock()

	return panel
}

func RevokePanel() {
	panelMu.Lock()
	defer panelMu.Unlock()

	panel = nil
}

func Setup() {
	username, password := vars.Get("PANEL_USERNAME"), vars.Get("PANEL_PASSWORD")
	url := vars.Get("PANEL_URL")
	newPanel := &Panel{
		client: resty.New(),
	}

	newPanel.client.SetBaseURL(url)

	if err := newPanel.Login(username, password); err != nil {
		panic(err)
	}

	panelMu.Lock()
	defer panelMu.Unlock()

	panel = newPanel
}
