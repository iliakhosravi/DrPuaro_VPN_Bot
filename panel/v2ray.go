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
	subLink, err := panel.SubLink(client)
	if err != nil {
		fmt.Println(err)
		return []string{}, err
	}

	return extractShortLinks(title, subLink)
}

// shortLinkConfig translates the parsed V2Ray configuration into a short link.
func ShortLinkConfigFromSubID(title, subID string) (string, error) {
	subLink, err := PanelSubLinkFromSubID(subID)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	configs, err := extractShortLinks(title, subLink)
	return configs[0], err
}

func extractShortLinks(title, subLink string) ([]string, error) {
	shortLinks := []string{}

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

func (panel *Panel) PanelSubLink(client Client) (string, error) {
	inboundClient, err := panel.GetInboundClient(&client)
	if err != nil {
		return "", err
	}
	return PanelSubLinkFromSubID(inboundClient.SubID)
}

func (panel *Panel) SubLink(client Client) (string, error) {
	inboundClient, err := panel.GetInboundClient(&client)
	if err != nil {
		return "", err
	}
	link, err := url.JoinPath(fmt.Sprintf("https://%s:%s/%s/%s", os.Getenv("SUB_URL"), os.Getenv("SUB_PORT"), os.Getenv("SUB_PATH"), inboundClient.SubID))
	if err != nil {
		return "", err
	}
	return link, nil
}

func PanelSubLinkFromSubID(subID string) (string, error) {
	link, err := url.JoinPath(fmt.Sprintf("https://%s:%s/%s/%s", os.Getenv("PANEL_SUB_URL"), os.Getenv("PANEL_SUB_PORT"), os.Getenv("PANEL_SUB_PATH"), subID))
	if err != nil {
		return "", err
	}
	return link, nil
}
