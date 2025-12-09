package sui

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
	"techybat.org/go-vpn/vars"
)

type Setting map[string]string

type Sui struct {
	client *resty.Client
}

func GetSui() *Sui {
	url := vars.Get("SUI_API_URL")
	token := vars.Get("SUI_API_TOKEN")
	client := resty.New().SetBaseURL(url)
	client.SetHeader("Token", token)
	if os.Getenv("env") == "dev" {
		client.SetDebug(true)
	}
	return &Sui{
		client: client,
	}
}

func (sui Sui) GetSetting() (Setting, error) {
	resp, err := sui.client.R().
		Get("/settings")

	if err != nil {
		return nil, err
	}

	parsed := gjson.Parse(resp.String())
	if !parsed.Get("success").Bool() {
		return nil, fmt.Errorf("VPN Panel Returned Error: %s\n", parsed.Get("msg").String())
	}

	settingStr := parsed.Get("obj").Raw
	setting := make(Setting)
	err = json.Unmarshal([]byte(settingStr), &setting)
	if err != nil {
		return nil, err
	}
	return setting, nil
}

func (sui Sui) GetSubUrl(client Client) (string, error) {
	setting, err := sui.GetSetting()
	if err != nil {
		return "", err
	}

	return url.JoinPath(setting["subURI"], client.Name)
}

func (sui Sui) LoadClients(lastUpdate, clientQuery string) ([]Client, error) {
	resp, err := sui.client.R().
		SetQueryParam("lu", lastUpdate).
		Get("/load")
	if err != nil {
		return nil, err
	}

	parsed := gjson.Parse(resp.String())
	if !parsed.Get("success").Bool() {
		return nil, fmt.Errorf("VPN Panel Returned Error: %s\n", parsed.Get("msg").String())
	}

	var clients []Client
	fullQuery := "obj.clients"
	if clientQuery != "" {
		fullQuery += "." + clientQuery
	}
	clientsStr := parsed.Get(fullQuery).Raw

	err = json.Unmarshal([]byte(clientsStr), &clients)
	if err != nil {
		return nil, err
	}

	return clients, nil
}

func (sui Sui) NewClient(client Client) (*Client, error) {
	return sui.PostClient(client, "new")
}

func (sui Sui) UpdateClient(client Client) (*Client, error) {
	return sui.PostClient(client, "edit")
}

func (sui Sui) PostClient(client Client, action string) (*Client, error) {
	payload, err := json.Marshal(client)
	if err != nil {
		return nil, err
	}

	resp, err := sui.client.R().
		SetFormData(map[string]string{
			"object": "clients",
			"action": action,
			"data":   string(payload),
		}).
		Post("/save")

	if err != nil {
		return nil, err
	}

	parsed := gjson.Parse(resp.String())
	if !parsed.Get("success").Bool() {
		return nil, fmt.Errorf("VPN Panel Returned Error: %s\n", parsed.Get("msg").String())
	}

	clientStr := parsed.Get(fmt.Sprintf(`obj.clients.#(name=="%s")`, client.Name)).Raw
	var retClient Client
	err = json.Unmarshal([]byte(clientStr), &retClient)
	if err != nil {
		return nil, err
	}

	return &retClient, nil
}

func (sui Sui) GetClients(query string) ([]Client, error) {
	resp, err := sui.client.R().
		Get("/clients")
	if err != nil {
		return nil, err
	}

	parsed := gjson.Parse(resp.String())
	if !parsed.Get("success").Bool() {
		return nil, fmt.Errorf("VPN Panel Returned Error: %s\n", parsed.Get("msg").String())
	}

	var clients []Client
	fullQuery := "obj.clients"
	if query != "" {
		fullQuery += "." + query
	}
	clientsStr := parsed.Get(fullQuery).Raw

	err = json.Unmarshal([]byte(clientsStr), &clients)
	if err != nil {
		return nil, err
	}

	return clients, nil
}

func (sui Sui) GetClientByName(name string) (*Client, error) {
	resp, err := sui.client.R().
		Get("/clients")
	if err != nil {
		return nil, err
	}

	parsed := gjson.Parse(resp.String())
	if !parsed.Get("success").Bool() {
		return nil, fmt.Errorf("VPN Panel Returned Error: %s\n", parsed.Get("msg").String())
	}

	var client Client
	clientStr := parsed.Get(fmt.Sprintf(`obj.clients.#(name=="%s")`, name)).Raw

	err = json.Unmarshal([]byte(clientStr), &client)
	if err != nil {
		return nil, err
	}

	return &client, nil
}

func (sui Sui) GetClientByID(id int) (*Client, error) {
	resp, err := sui.client.R().
		SetQueryParam("id", fmt.Sprint(id)).
		Get("/clients")

	if err != nil {
		return nil, err
	}

	parsed := gjson.Parse(resp.String())
	if !parsed.Get("success").Bool() {
		return nil, fmt.Errorf("VPN Panel Returned Error: %s\n", parsed.Get("msg").String())
	}

	var client Client
	clientStr := parsed.Get(fmt.Sprintf(`obj.clients.#(id==%d)`, id)).Raw
	err = json.Unmarshal([]byte(clientStr), &client)
	if err != nil {
		fmt.Printf("ATTENTION! ATTENTION!\n\n%s\n\nTHANK YOU!\n", clientStr)
		return nil, err
	}

	return &client, nil
}

func (sui Sui) GetShortLinks(clientID int) ([]string, error) {
	client, err := sui.GetClientByID(clientID)
	if err != nil {
		return nil, err
	}

	var shortLinks []string
	for _, link := range client.Links {
		shortLinks = append(shortLinks, link["uri"])
	}

	return shortLinks, nil
}
