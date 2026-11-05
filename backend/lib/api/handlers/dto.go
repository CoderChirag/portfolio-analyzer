package handlers

type SignUpBody struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	PAN   string `json:"pan"`
}

type SignupResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
