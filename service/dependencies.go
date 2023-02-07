package service

import (
	"github.com/Coderx44/MovieTicketingPortal/app"
	"github.com/Coderx44/MovieTicketingPortal/booking"
	"github.com/Coderx44/MovieTicketingPortal/db"
)

type dependencies struct {
	BookingService booking.Service
}

func initDependencies() (dependencies, error) {
	logger := app.GetLogger()

	appDB := app.GetDB()
	dbStore := db.NewStorer(appDB)

	bookingService := booking.NewBookingService(dbStore, logger)

	return dependencies{
		BookingService: bookingService,
	}, nil
}
