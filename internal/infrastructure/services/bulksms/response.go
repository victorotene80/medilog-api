package bulksms

type SendMessageResponse struct {
	Status  string               `json:"status"`
	Code    string               `json:"code"`
	Message string               `json:"message"`
	Data    SendMessageResponseData `json:"data"`
}

type SendMessageResponseData struct {
	MessageID       string  `json:"message_id"`
	Cost            float64 `json:"cost"`
	Currency        string  `json:"currency"`
	RecipientsCount int     `json:"recipients_count"`
	GatewayUsed     string  `json:"gateway_used"`
	SandboxMode     bool    `json:"sandbox_mode"`
}
