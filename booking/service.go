package booking

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Coderx44/MovieTicketingPortal/db"
	"go.uber.org/zap"
)

var secretKey = []byte("ThisIsMyFistGolangProjecT")

const DateOnly = "2006-01-02"

type Service interface {
	CreateNewUser(ctx context.Context, u NewUser) (user_id uint, err error)
	Login(ctx context.Context, authU Authentication) (tokenString string, tokenExpirationTime time.Time, err error)
	AddMovie(ctx context.Context, m NewMovie) (movie_id uint, err error)
	AddScreen(ctx context.Context, s NewScreen) (screen_id uint, err error)
	AddMultiplex(ctx context.Context, m NewMultiplex) (multiplex_id uint, err error)
	AddLocation(ctx context.Context, l NewLocation) (location_id uint, err error)
	AddShow(ctx context.Context, s NewShow) (show_id uint, err error)
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
	tokenString, tokenExpirationTime, err = generateJWT(user.Email, user.Role)
	if err != nil {
		return
	}
	return

}

func (b *bookingService) AddMovie(ctx context.Context, m NewMovie) (movie_id uint, err error) {
	rDate, errr := time.Parse(DateOnly, m.Release_date)
	if errr != nil {
		err = errors.New("err: failed to add movie")
		return
	}
	newM := db.Movie{
		Title:        m.Title,
		Language:     m.Language,
		Release_date: rDate,
		Genre:        m.Genre,
		Duration:     m.Duration,
	}

	movie_id, err = b.store.AddMovie(ctx, newM)
	if err != nil {
		b.logger.Errorf("Err: Adding Movie: %v", err.Error())
		return
	}

	b.logger.Infof("Movie id  %v", movie_id)

	return

}

func (b *bookingService) AddScreen(ctx context.Context, s NewScreen) (screen_id uint, err error) {

	newSn := db.Screen{
		Screen_number:    s.Screen_number,
		Total_seats:      s.Total_seats,
		Sound_system:     s.Sound_system,
		Screen_dimension: s.Screen_dimension,
		Multiplex_id:     s.Multiplex_id,
	}

	if ok := MultiplexIdExists(b, ctx, newSn.Multiplex_id); !ok {
		log.Println(err)
		err = errors.New("err: invalid Multiplex id")
		return
	}
	log.Println("newsn", newSn)
	screen_id, err = b.store.AddScreen(ctx, newSn)
	if err != nil {
		b.logger.Errorf("Err: Adding Screen: %v", err.Error())
		return
	}

	b.logger.Infof("Screen ID  %v", screen_id)

	return

}

func (b *bookingService) AddLocation(ctx context.Context, l NewLocation) (location_id uint, err error) {

	newL := db.Location{
		City:    l.City,
		State:   l.State,
		Pincode: l.Pincode,
	}

	location_id, err = b.store.AddLocation(ctx, newL)
	if err != nil {
		b.logger.Errorf("Err: Adding Location: %v", err.Error())
		return
	}

	b.logger.Infof("Location ID  %v", location_id)
	return
}

func (b *bookingService) AddMultiplex(ctx context.Context, m NewMultiplex) (multiplex_id uint, err error) {

	location_id, err := getLocationID(ctx, b, m.City, m.State, m.Pincode)
	if err != nil {
		b.logger.Errorf("Err: Adding Multiplex: %v", err.Error())
		return
	}

	newM := db.Multiplexe{
		Name:          m.Name,
		Contact:       m.Contact,
		Total_screens: m.Total_screens,
		Locality:      m.Locality,
		Location_id:   int(location_id),
	}

	multiplex_id, err = b.store.AddMultiplex(ctx, newM)
	if err != nil {
		b.logger.Errorf("Err: Adding Multiplex: %v", err.Error())
		return
	}

	b.logger.Infof("Multiplex ID  %v", multiplex_id)
	return

}

func (b *bookingService) AddShow(ctx context.Context, s NewShow) (show_id uint, err error) {
	log.Println("Show", s)
	if ok := MultiplexIdExists(b, ctx, s.Multiplex_id); !ok {
		log.Println(err)
		err = errors.New("err: invalid Multiplex id")
		return
	}
	// var screen db.Screen

	screen, ok := ScreenExists(b, ctx, s.Screen, s.Multiplex_id)
	if !ok {
		log.Println(err)
		err = errors.New("err: invalid screen number")
		return
	}
	s.Screen_id = screen.Screen_id
	movie_id, ok := MovieExists(b, ctx, s.Movie)
	if !ok {
		log.Println(err)
		err = errors.New("err: Movie doesn't exist")
		return
	}
	s.Movie_id = movie_id
	rDate, err := time.Parse(DateOnly, s.Date)
	if err != nil {
		err = errors.New("invalid date format")
		return
	}
	st_time, err := time.Parse(time.Kitchen, strings.TrimSpace(s.Start_time))
	if err != nil {
		log.Println(err)
	}
	end_time, _ := time.Parse(time.Kitchen, s.End_time)
	if err != nil {
		log.Println(err)
	}
	newSh := db.Show{
		Show_date:    rDate,
		Start_time:   st_time,
		End_time:     end_time,
		Screen_id:    s.Screen_id,
		Movie_id:     s.Movie_id,
		Multiplex_id: s.Multiplex_id,
	}
	log.Println("newsh", newSh)
	show_id, err = b.store.AddShow(ctx, newSh)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			err = errors.New("err : overlapping sow times")
		}
		b.logger.Errorf("Err: Adding Show: %v", err.Error())
		return
	}

	err = b.store.AddSeats(ctx, screen.Total_seats, int(show_id))
	if err != nil {
		b.logger.Errorf("Err: Adding seats for show: %v", err.Error())
		return
	}
	b.logger.Infof("Show ID  %v", show_id)

	return

}
