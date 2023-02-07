CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE TABLE IF NOT EXISTS users(

user_id  SERIAL PRIMARY KEY,
name text,
email citext UNIQUE,
password text,
phone_number text

);

CREATE TABLE IF NOT EXISTS movies(

movie_id SERIAL PRIMARY KEY,
title text,
language text,
poster bytea,
release_date date,
genre text,
duration text

);


CREATE TABLE IF NOT EXISTS locations(
    location_id SERIAL PRIMARY KEY,
    city text,
    state text,
    pincode int

);


CREATE TABLE IF NOT EXISTS multiplexes(
multiplex_id SERIAL PRIMARY KEY,
name text,
contact text,
total_screens int,
locality text,
location_id int REFERENCES locations (location_id)
);

CREATE TABLE IF NOT EXISTS screens(

    screen_id SERIAL PRIMARY KEY,
    total_seats int,
    sound_system text,
    screen_dimension text,
    multiplex_id int REFERENCES multiplexes (multiplex_id)

);

CREATE TABLE IF NOT EXISTS shows(

show_id SERIAL PRIMARY KEY,
show_date date,
start_time time,
end_time time,
screen_id int REFERENCES screens (screen_id),
movie_id int REFERENCES movies (movie_id),
multiplex_id int REFERENCES multiplexes (multiplex_id)

);
 

CREATE TABLE IF NOT EXISTS screen_types(

screen_type_id SERIAL PRIMARY KEY,
seat_number int,
class text,
screen_id int REFERENCES screens (screen_id)

);

CREATE TABLE IF NOT EXISTS seats(

seat_id SERIAL PRIMARY KEY,
price int,
status text,
screen_type_id int REFERENCES screen_types (screen_type_id),
show_id int REFERENCES shows (show_id)
);

CREATE TABLE IF NOT EXISTS bookings(

booking_id SERIAL PRIMARY KEY,
seat_no int,
status text,
expiry timestamp,
user_id int REFERENCES users (user_id),
seat_id int REFERENCES seats (seat_id),
show_id int REFERENCES shows (show_id)

);

CREATE TABLE IF NOT EXISTS transactions(

transaction_id SERIAL PRIMARY KEY,
price int,
time_stamp timestamp,
booking_id int REFERENCES bookings (booking_id)
); 