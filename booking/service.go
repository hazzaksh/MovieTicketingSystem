package booking

import (
	"context"
	"database/sql"
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
	Login(ctx context.Context, authU Authentication) (tokenString string, err error)
	AddMovie(ctx context.Context, m NewMovie) (movie_id uint, err error)
	AddScreen(ctx context.Context, s NewScreen) (screen_id uint, err error)
	AddMultiplex(ctx context.Context, m NewMultiplex) (multiplex_id uint, err error)
	AddLocation(ctx context.Context, l NewLocation) (location_id int, err error)
	AddShow(ctx context.Context, s NewShow) (show_id uint, err error)
	GetAllMultiplexesByCity(ctx context.Context, city string) (m []NewMultiplex, err error)
	GetAllShowsByDateAndMultiplexId(ctx context.Context, date string, multiplex_id int) (map[string][]MultiplexShow, error)
	GetAllShowsByMovieAndDate(ctx context.Context, date string, title string, city string) (map[string][]MultiplexShow, error)
	GetAllSeatsByShowID(ctx context.Context, show_id int) (map[int][]Seats, error)
	AddBookingsBySeatId(ctx context.Context, seats []int, email string) (invoice Invoice, err error)
	GetUpcomingMovies(ctx context.Context, date string) (m []NewMovie, err error)
	GetMovieByTitle(ctx context.Context, title string) (m NewMovie, err error)
	CancelBooking(ctx context.Context, id int) (err error)
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
		PhoneNumber: u.PhoneNumber,
		Role:        u.Role,
	}

	user_id, err = b.store.CreateUser(ctx, newU)
	if err != nil {
		// b.logger.Errorf("Err creating user account: %v", err)
		if err.Error() == "pq: duplicate key value violates unique constraint \"users_email_key\"" {
			err = fmt.Errorf("user exists for the given email")
			return
		}

		// b.logger.Infof("User Details  %v", user_id)

		return
	}

	return
}

func (b *bookingService) Login(ctx context.Context, authU Authentication) (tokenString string, err error) {
	user, err := b.store.GetUserByEmail(ctx, authU.Email)

	if err != nil && err.Error() == "user does not exist in db" {
		err = errors.New("unauthorized")
		return
	} else if err != nil {
		return
	}

	check := CheckPasswordHash(authU.Password, user.Password)
	if !check {
		err = errors.New("unauthorized")
		return
	}
	tokenString, err = generateJWT(user.Email, user.Role)
	if err != nil {
		return
	}
	return

}

func (b *bookingService) AddMovie(ctx context.Context, m NewMovie) (movie_id uint, err error) {
	// rDate, errr := time.Parse(DateOnly, m.Release_date)

	newM := db.Movie{
		Title:        m.Title,
		Language:     m.Language,
		Release_date: m.Release_date,
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

func (b *bookingService) AddLocation(ctx context.Context, l NewLocation) (location_id int, err error) {

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
	movie, err := MovieExists(b, ctx, s.Movie)
	if err != nil && err.Error() == "movie doesn't exist" {
		log.Println(err)
		err = errors.New("err: Movie doesn't exist")
		return
	} else if err != nil {
		return
	}

	s.Movie_id = movie.Movie_id
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

	// log.Println("newsh", newSh)
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

func (b *bookingService) GetAllMultiplexesByCity(ctx context.Context, city string) (m []NewMultiplex, err error) {

	location, err := b.store.GetLocationIdByCity(ctx, city)
	if err != nil {
		// b.logger.Errorf("Err: Fetching All multiplexes: %v", err.Error())
		err = errors.New("failed to get multiplexes")
		return
	}

	multiplexes, err := b.store.GetAllMultiplexesByLocationID(ctx, location.Location_id)
	if err != nil {
		// b.logger.Errorf("Err: Fetching All multiplexes: %v", err.Error())
		err = errors.New("failed to get multiplexes")
		return
	}

	for _, multiplx := range multiplexes {

		multiplex := NewMultiplex{
			multiplx.Multiplex_id,
			multiplx.Name,
			multiplx.Contact,
			multiplx.Total_screens,
			multiplx.Locality,
			location.City,
			location.State,
			location.Pincode,
			location.Location_id,
		}

		m = append(m, multiplex)
	}

	return

}

func (b *bookingService) GetAllShowsByDateAndMultiplexId(ctx context.Context, date string, multiplex_id int) (map[string][]MultiplexShow, error) {
	shows := make(map[string][]MultiplexShow)
	log.Println(date)
	cDate, err := time.Parse("2006-01-02", strings.TrimSpace(date))
	log.Println(err)
	if err != nil {
		err = errors.New("invalid date format")
		return shows, err
	}

	allShows, err := b.store.GetAllShowsByDateAndMultiplexId(ctx, cDate, multiplex_id)
	if err != nil {
		b.logger.Errorf("Err: Fetching All shows: %v", err.Error())
		return shows, err
	}

	for _, value := range allShows {
		shows[value.Title] = append(shows[value.Title], MultiplexShow(value))
	}

	return shows, err

}

func (b *bookingService) GetAllShowsByMovieAndDate(ctx context.Context, date string, title string, city string) (map[string][]MultiplexShow, error) {
	shows := make(map[string][]MultiplexShow)
	log.Println(date)
	cDate, err := time.Parse("2006-01-02", strings.TrimSpace(date))
	log.Println(err)
	if err != nil {
		err = errors.New("invalid date format")
		return shows, err
	}

	allShows, err := b.store.GetAllShowsByMovieAndDate(ctx, title, city, cDate)
	if err != nil {
		// b.logger.Errorf("Err: Fetching All shows: %v", err.Error())
		return shows, err
	}

	for _, value := range allShows {
		shows[value.Multiplex_name+" "+value.Locality] = append(shows[value.Title], MultiplexShow(value))
	}

	return shows, err

}

func (b *bookingService) GetAllSeatsByShowID(ctx context.Context, show_id int) (map[int][]Seats, error) {

	seats := make(map[int][]Seats)

	allSeats, err := b.store.GetSeatsByShowID(ctx, show_id)

	if err != nil {
		// b.logger.Errorf("Err: Fetching Seats: %v", err.Error())
		return seats, err
	}
	for _, value := range allSeats {
		seats[value.Seat_number] = append(seats[value.Seat_number], Seats(value))
	}

	return seats, err

}

func createInvoice(b *bookingService, ctx context.Context, show_id int) (invoice Invoice, err error) {

	invoiceDb, err := b.store.GetInvoiceDetails(ctx, show_id)
	return Invoice(invoiceDb), err
}

func (b *bookingService) AddBookingsBySeatId(ctx context.Context, seats []int, email string) (invoice Invoice, err error) {

	log.Println("in service", seats)
	available, err := b.store.CheckAvailability(ctx, seats)
	if err == sql.ErrNoRows {
		err = errors.New("Seats not available")
		return
	}
	if !available {
		err = errors.New("Seats not available")
		return
	}
	seat, err := b.store.GetSeatsByID(ctx, seats)
	if err != nil {
		return
	}
	show_id := seat[0].Show_id
	var seat_num []int
	for _, value := range seat {
		seat_num = append(seat_num, value.Seat_number)
	}

	err = b.store.AddBookingsBySeatId(ctx, seats, email, show_id, seat_num)
	if err != nil {
		return
	}

	invoice, err = createInvoice(b, ctx, show_id)
	invoice.Email = email
	invoice.Seats = seat_num
	invoice.Total_price = len(seat_num) * seat[0].Price
	if err != nil {
		err = errors.New("err: cannot generate invoice")
		return
	}
	return
}

func (b *bookingService) GetUpcomingMovies(ctx context.Context, date string) (m []NewMovie, err error) {

	movies, err := b.store.GetUpcomingMovies(ctx, date)
	if err != nil {
		return []NewMovie{}, err
	}

	for _, value := range movies {
		m = append(m, NewMovie(value))
	}

	return
}

func (b *bookingService) GetMovieByTitle(ctx context.Context, title string) (m NewMovie, err error) {

	m, err = MovieExists(b, ctx, title)
	if err != nil {
		err = errors.New("movie not found")
		return
	}
	return
}

func (b *bookingService) CancelBooking(ctx context.Context, id int) (err error) {
	err = b.store.DeleteByBookingByID(ctx, id)
	log.Println(err)
	if err != nil {
		err = errors.New("err: cannot cancel booking")
		return
	}
	return
}
