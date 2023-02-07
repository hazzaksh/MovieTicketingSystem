package booking

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Coderx44/MovieTicketingPortal/db"
	"go.uber.org/zap"
)

var secretKey = []byte("ThisIsMyFistGolangProjecT")

type Service interface {
	CreateNewUser(ctx context.Context, u NewUser) (user_id uint, err error)
	Login(ctx context.Context, authU Authentication) (tokenString string, tokenExpirationTime time.Time, err error)
}

type bookingService struct {
	store  db.Storer
	logger *zap.SugaredLogger
}

func NewBookingService(s db.Storer, l *zap.SugaredLogger) Service {
	return &bookingService{
		store:  s,
		logger: l,
	}
}

func (b *bookingService) CreateNewUser(ctx context.Context, u NewUser) (user_id uint, err error) {

	newU := db.User{
		Name:        u.Name,
		Email:       u.Email,
		Password:    u.Password,
		PhoneNumber: u.Phone_number,
		Role:        u.Role,
	}

	user_id, err = b.store.CreateUser(ctx, newU)
	if err != nil {
		b.logger.Errorf("Err creating user account: %v", err.Error())
		if err.Error() == "pq: duplicate key value violates unique constraint \"users_email_key\"" {
			err = fmt.Errorf("user exists for the given email")
			return
		}

		b.logger.Infof("User Details  %v", user_id)

		return
	}

	return
}

func (b *bookingService) Login(ctx context.Context, authU Authentication) (tokenString string, tokenExpirationTime time.Time, err error) {
	user, err := b.store.GetUserByEmail(ctx, authU.Email)

	if err == errors.New("user does not exist in db") {
		err = errors.New("unauthorized")
		return
	}

	check := CheckPasswordHash(authU.Password, user.Password)
	if !check {
		err = errors.New("username or password is incorrect")
		return
	}
	tokenString, tokenExpirationTime, err = generateJWT(authU.Email, "user")
	if err != nil {
		return
	}
	return

}
