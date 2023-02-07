package booking

type NewUser struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	Phone_number string `json:"phone_number"`
}

type NewUserResponse struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	Phone_number string `json:"phone_number"`
}
