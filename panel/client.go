package panel

import (
	"encoding/json"
	"fmt"
)

const (
	ONE_GB             = 1073741824
	CLIENT_PATH        = "/panel/api/inbounds/getClientTraffics/"
	ADD_CLIENT_PATH    = "/panel/api/inbounds/addClient"
	UPDATE_CLIENT_PATH = "/panel/api/inbounds/updateClient/"
)

type Client struct {
	ID         int    `json:"id"`
	InboundID  int    `json:"inboundId"`
	Enable     bool   `json:"enable"`
	Email      string `json:"email"`
	ExpiryTime int64  `json:"expiryTime"`
	Total      int    `json:"total"`
	Up         int    `json:"up"`
	Down       int    `json:"down"`
	Reset      int    `json:"reset"`
}

type ClientForm struct {
	ID         string `json:"id"`
	AlterID    int    `json:"alterId"`
	Email      string `json:"email"`
	LimitIP    int    `json:"limitIp"`
	TotalGB    int64  `json:"totalGB"`
	ExpiryTime int64  `json:"expiryTime"`
	Enable     bool   `json:"enable"`
	TgID       string `json:"tgId"`
	SubID      string `json:"subId"`
}

type ClientResponse struct {
	BasicResponse
	Object Client `json:"obj"`
}

type addClientPayload struct {
	ID       int    `json:"id"`
	Settings string `json:"settings"`
}

type addClientSettings struct {
	Clients []ClientForm `json:"clients"`
}

func (panel *Panel) GetClient(email string) (Client, error) {
	var res ClientResponse
	_, err := panel.client.R().
		SetResult(&res).
		Get(CLIENT_PATH + email)

	if err != nil {
		fmt.Println("error: getting client failed.\nerr:", err)
		return Client{}, fmt.Errorf("getting client failed.\nerr:%s", err)
	}

	if !res.Success {
		fmt.Println("error: server returned no client. message: ", res.Message)
		return Client{}, fmt.Errorf("error: server returned no client. message: %s", res.Message)
	}

	return res.Object, nil
}

func (panel *Panel) AddClient(inboundID int, clientForm ClientForm) (Client, error) {
	clientBytes, _ := json.Marshal(addClientSettings{
		Clients: []ClientForm{clientForm},
	})
	payload := addClientPayload{
		ID:       inboundID,
		Settings: string(clientBytes),
	}

	var res BasicResponse
	resp, err := panel.client.R().
		EnableTrace().
		SetBody(payload).
		SetResult(&res).
		Post(ADD_CLIENT_PATH)

	curlCmdExecuted := resp.Request.GenerateCurlCommand()

	// Explore curl command
	fmt.Println("Curl Command:\n  ", curlCmdExecuted+"\n")
	if err != nil {
		return Client{}, fmt.Errorf("error: adding client failed.\nerr:%s", err)
	}

	if !res.Success {
		fmt.Println(res)
		return Client{}, fmt.Errorf("error: server couldn't add client. message: %s", res.Message)
	}

	return panel.GetClient(clientForm.Email)
}

func (panel *Panel) UpdateClient(inboundID int, clientForm ClientForm) (Client, error) {
	clientBytes, _ := json.Marshal(addClientSettings{
		Clients: []ClientForm{clientForm},
	})
	payload := addClientPayload{
		ID:       inboundID,
		Settings: string(clientBytes),
	}

	var res BasicResponse
	_, err := panel.client.R().
		SetBody(payload).
		SetResult(&res).
		Post(UPDATE_CLIENT_PATH + clientForm.ID)
	if err != nil {
		return Client{}, fmt.Errorf("error: update client failed.\nerr:%s", err)
	}

	if !res.Success {
		return Client{}, fmt.Errorf("error: server couldn't update client. message: %s", res.Message)
	}

	return panel.GetClient(clientForm.Email)
}

func (panel *Panel) StoreClient(inboundID int, clientForm ClientForm) (Client, error) {
	client, err := panel.GetClient(clientForm.Email)
	if err != nil || (client == Client{}) {
		return panel.AddClient(inboundID, clientForm)
	}
	fmt.Println("STORE CLIENT GETCLIENT: ", client, err)
	return panel.UpdateClient(inboundID, clientForm)
}

func (panel *Panel) ResetClientStats(inboundID int, email string) (Client, error) {
	var res BasicResponse
	_, err := panel.client.R().
		SetResult(&res).
		Post(fmt.Sprintf("/panel/api/inbounds/%d/resetClientTraffic/%s", inboundID, email))

	if err != nil {
		return Client{}, fmt.Errorf("error: reset client stats failed.\nerr:%s", err)
	}

	if !res.Success {
		return Client{}, fmt.Errorf("error: server couldn't reset client stats. message: %s", res.Message)
	}

	return panel.GetClient(email)
}

// func generateClientPayload(uuid, email, tgId string, totalGB, limitIP int, expiryTime int64) string {
// 	return fmt.Sprintf("{\"clients\":[{\"id\":\"%s\",\"alterId\":0,\"email\":\"%s\",\"limitIp\":%d,\"totalGB\":%d,\"expiryTime\":%d,\"enable\":true,\"tgId\":\"%s\",\"subId\":\"\"}]}", uuid, email, limitIP, totalGB, expiryTime, tgId)
// }
