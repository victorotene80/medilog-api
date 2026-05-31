package twilio

type SendMessageResult struct {
	SID          string `json:"sid"`
	AccountSID   string `json:"account_sid"`
	From         string `json:"from"`
	To           string `json:"to"`
	Body         string `json:"body"`
	Status       string `json:"status"`
	ErrorCode    *int   `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}

type twilioErrorResponse struct {
	Code     int    `json:"code"`
	Message  string `json:"message"`
	MoreInfo string `json:"more_info"`
	Status   int    `json:"status"`
}