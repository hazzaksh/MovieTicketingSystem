# syntax=docker/dockerfile:1

## Build
FROM golang:1.19-buster AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN go build -o /movieticket-build-file

## Deploy
FROM alpine

WORKDIR /

COPY --from=build /movieticket-build-file /movieticket-build-file

EXPOSE 3000


CMD ["/movieticket-build-file", "start"]
