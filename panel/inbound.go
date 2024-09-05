package panel

import (
	"encoding/json"
	"fmt"
)

const (
	INBOUNDS_PATH    = "/panel/api/inbounds/list"
	GET_INBOUND_PATH = "/panel/api/inbounds/get/"
)

type Inbound struct {
	ID                int             `json:"id"`
	Tag               string          `json:"tag"`
	Protocol          string          `json:"protocol"`
	Remark            string          `json:"remark"`
	Enable            bool            `json:"enable"`
	Port              int             `json:"port"`
	StrSettings       string          `json:"settings"`
	StrStreamSettings string          `json:"streamSettings"`
	Settings          InboundSettings `json:"-"`
	StreamSettings    StreamSettings  `json:"-"`
}

type InboundClient struct {
	ID         string `json:"id"`
	Flow       string `json:"flow"`
	Email      string `json:"email"`
	LimitIP    int    `json:"limitIp"`
	TotalGB    int    `json:"totalGB"`
	ExpiryTime int64  `json:"expiryTime"`
	Enable     bool   `json:"enable"`
	TgID       string `json:"tgId"`
	SubID      string `json:"subId"`
	Reset      int    `json:"reset"`
}

type StreamSettings struct {
	Network     string `json:"network"`
	Security    string `json:"security"`
	TLSSettings struct {
		ServerName string   `json:"serverName"`
		MinVersion string   `json:"minVersion"`
		MaxVersion string   `json:"maxVersion"`
		ALPN       []string `json:"alpn"`
		Settings   struct {
			AllowInsecure bool   `json:"allowInsecure"`
			Fingerprint   string `json:"fingerprint"`
		} `json:"settings"`
	} `json:"tlsSettings"`
	WSSettings struct {
		Path                string            `json:"path"`
		AcceptProxyProtocol bool              `json:"acceptProxyProtocol"`
		Host                string            `json:"host"`
		Headers             map[string]string `json:"headers"`
	} `json:"wsSettings"`
	Description string `json:"description"`
}

type InboundSettings struct {
	Clients    []InboundClient `json:"clients"`
	Decryption string          `json:"decryption"`
	Fallbacks  []string        `json:"fallbacks"`
}

func (panel *Panel) GetInbounds() ([]Inbound, error) {
	type inboundsResponse struct {
		BasicResponse
		Object []Inbound `json:"obj"`
	}
	var insRes inboundsResponse
	resp, err := panel.client.R().
		EnableTrace().
		SetResult(&insRes).
		Get(INBOUNDS_PATH)

	curlCmdExecuted := resp.Request.GenerateCurlCommand()
	fmt.Println("Curl Command:\n  ", curlCmdExecuted+"\n")

	fmt.Println("Err: ", err)
	if err != nil {
		return []Inbound{}, fmt.Errorf("error: getting inbounds failed.\nerr:%s", err)
	}

	fmt.Println("Message: ", insRes.Message)
	if !insRes.Success {
		return []Inbound{}, fmt.Errorf("error: getting inbounds failed. response msg:%s", insRes.Message)
	}

	inbounds := insRes.Object
	fmt.Println("inbounds: ", inbounds)

	for _, inbound := range inbounds {
		json.Unmarshal([]byte(inbound.StrSettings), &inbound.Settings)
		json.Unmarshal([]byte(inbound.StrStreamSettings), &inbound.StreamSettings)
	}

	fmt.Println("\n\n\n\n\n\nafter marshal inbounds: ", inbounds)

	return inbounds, nil
}

func (panel *Panel) GetInbound(inboundID int) (Inbound, error) {
	type inboundResponse struct {
		BasicResponse
		Object Inbound `json:"obj"`
	}

	var insRes inboundResponse
	_, err := panel.client.R().
		SetResult(&insRes).
		Get(fmt.Sprintf("%s%d", GET_INBOUND_PATH, inboundID))
	if err != nil {
		return Inbound{}, fmt.Errorf("error: getting inbound failed.\nerr:%s", err)
	}

	if !insRes.Success {
		return Inbound{}, fmt.Errorf("error: getting inbound failed. response msg:%s", insRes.Message)
	}

	inbound := insRes.Object
	json.Unmarshal([]byte(inbound.StrSettings), &inbound.Settings)
	json.Unmarshal([]byte(inbound.StrStreamSettings), &inbound.StreamSettings)

	return inbound, nil
}

func (inbound *Inbound) FindClientFromEmail(email string) (*InboundClient, error) {
	for _, inClient := range inbound.Settings.Clients {
		if email == inClient.Email {
			return &inClient, nil
		}
	}
	return &InboundClient{}, fmt.Errorf("client not found in this inbound")
}
