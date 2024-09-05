package panel

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
)

func (panel *Panel) GetInboundClient(client *Client) (*InboundClient, error) {
	inbound, _ := panel.GetInbound(client.InboundID)
	return inbound.FindClientFromEmail(client.Email)
}

// shortLinkConfig translates the parsed V2Ray configuration into a short link.
func (panel *Panel) ShortLinksConfig(title string, client Client) ([]string, error) {
	shortLinks := []string{}

	subLink, err := panel.SubLink(client)
	if err != nil {
		fmt.Println(err)
		return shortLinks, err
	}

	resp, err := resty.New().
		R().
		Get(subLink)

	if err != nil {
		fmt.Println(err)
		return shortLinks, err
	}

	body := string(resp.Body())
	shortBytes, err := base64.StdEncoding.DecodeString(body)

	if err != nil {
		fmt.Println(err)
		return shortLinks, err
	}

	shortLinksStr := string(shortBytes)
	for _, shortLink := range strings.Split(shortLinksStr, "\n") {
		shortLink = strings.Split(shortLink, "#")[0] + "#" + title
		shortLinks = append(shortLinks, shortLink)
	}

	return shortLinks, nil
}

func (panel *Panel) SubLink(client Client) (string, error) {
	inboundClient, err := panel.GetInboundClient(&client)
	if err != nil {
		return "", err
	}
	link, err := url.JoinPath(fmt.Sprintf("https://%s:%s/%s/%s", os.Getenv("PANEL_SUB_URL"), os.Getenv("PANEL_SUB_PORT"), os.Getenv("PANEL_SUB_PATH"), inboundClient.SubID))
	if err != nil {
		return "", err
	}
	return link, nil
}
