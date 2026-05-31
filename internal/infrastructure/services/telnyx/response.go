package telnyx

type SendMessageResult struct {
	ID                 string `json:"id"`
	RecordType         string `json:"record_type"`
	Direction          string `json:"direction"`
	From               string `json:"from"`
	To                 string `json:"to"`
	Text               string `json:"text"`
	Type               string `json:"type"`
	OrganizationID     string `json:"organization_id"`
	MessagingProfileID string `json:"messaging_profile_id"`
	Status             string `json:"carrier_status"`
}

type telnyxErrorDetail struct {
	Code   string `json:"code"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
	Source struct {
		Pointer string `json:"pointer"`
	} `json:"source"`
}

type telnyxErrorResponse struct {
	Errors []telnyxErrorDetail `json:"errors"`
}

type telnyxDataWrapper struct {
	Data SendMessageResult `json:"data"`
}
