package booking

import (
	"github.com/Coderx44/MovieTicketingPortal/db"
	"go.uber.org/zap"
)

type Service interface {
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
