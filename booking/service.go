package booking

import (
	"context"
	"fmt"

	"github.com/Coderx44/MovieTicketingPortal/db"
	"go.uber.org/zap"
)

type Service interface {
	CreateNewUser(ctx context.Context, u NewUser) (user_id uint, err error)
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
