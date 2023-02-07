package booking

import "github.com/golang-jwt/jwt"

type NewUser struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	Phone_number string `json:"phone_number"`
	Role         string `json:"role"`
}

type NewUserResponse struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	Phone_number string `json:"phone_number"`
}

type Authentication struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Claims struct {
	Role  string `json:"role"`
	Email string `json:"email"`
	jwt.StandardClaims
}

type LoginResp struct {
	Token string `json:"token"`
	Mssg  string `json:"message"`
}
