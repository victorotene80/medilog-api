package response

type CountryResponse struct {
	Code     string  `json:"code"`
	Name     string  `json:"name"`
	DialCode string  `json:"dial_code"`
	Flag     *string `json:"flag,omitempty"`
}
