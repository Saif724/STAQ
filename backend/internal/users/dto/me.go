package dto

type MeResponse struct {
	ID            string `json:"id"`
	FullName      string `json:"full_name"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}
