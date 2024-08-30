package panel

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

// shortLinkConfig translates the parsed V2Ray configuration into a short link.
func (panel *Panel) ShortLinkConfig(description string, client Client) string {
	inbound, _ := panel.GetInbound(client.InboundID)
	streamSt := inbound.StreamSettings
	inboundClient, _ := inbound.FindClientFromEmail(client.Email)
	// Construct the protocol-specific fields
	protocolFields := fmt.Sprintf("type=%s", streamSt.Network)

	// Construct the transport-specific fields (assuming WebSocket for this example)
	transportFields := fmt.Sprintf("path=%s", url.QueryEscape(streamSt.WSSettings.Path))
	if streamSt.WSSettings.Host != "" {
		transportFields += fmt.Sprintf("&host=%s", streamSt.WSSettings.Host)
	}
	if streamSt.WSSettings.AcceptProxyProtocol {
		transportFields += "&acceptProxyProtocol=true"
	}
	for key, value := range streamSt.WSSettings.Headers {
		transportFields += fmt.Sprintf("&header-%s=%s", url.QueryEscape(key), url.QueryEscape(value))
	}

	// Construct the TLS-specific fields
	tlsFields := fmt.Sprintf("security=%s", streamSt.Security)
	if streamSt.Security == "tls" {
		tlsFields += fmt.Sprintf("&sni=%s", streamSt.TLSSettings.ServerName)
		// tlsFields += fmt.Sprintf("&minVersion=%s", streamSt.TLSSettings.MinVersion)
		// tlsFields += fmt.Sprintf("&maxVersion=%s", streamSt.TLSSettings.MaxVersion)
		if len(streamSt.TLSSettings.ALPN) > 0 {
			tlsFields += fmt.Sprintf("&alpn=%s", url.QueryEscape(strings.Join(streamSt.TLSSettings.ALPN, ",")))
		}
		tlsFields += fmt.Sprintf("&allowInsecure=%v", streamSt.TLSSettings.Settings.AllowInsecure)
		tlsFields += fmt.Sprintf("&fp=%s", streamSt.TLSSettings.Settings.Fingerprint)
		tlsFields += fmt.Sprintf("&flow=%s", inboundClient.Flow)
	}

	// Assemble the short link
	shortLink := fmt.Sprintf(
		"%s://%s@%s:%d?%s&%s&%s#%s",
		inbound.Protocol, // The protocol (vless, vmess, etc.)
		inboundClient.ID,
		os.Getenv("CONFIG_HOST"),
		inbound.Port,
		protocolFields,
		transportFields,
		tlsFields,
		url.QueryEscape(description),
	)

	return shortLink
}
