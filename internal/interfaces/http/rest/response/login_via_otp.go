package response

type LoginViaOTPResponse struct {
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	User         UserResponse  `json:"user"`
	ExpiresIn    int64         `json:"expires_in"`
}
