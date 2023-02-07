package service

import (
	"fmt"
	"strconv"

	"github.com/Coderx44/MovieTicketingPortal/config"
	"github.com/urfave/negroni"
)

func StartApiServer() {
	port := config.AppPort()
	server := negroni.Classic()

	dependencies, err := initDependencies()
	if err != nil {
		panic(err)
	}

	router := initRouter(dependencies)
	server.UseHandler(router)

	addr := fmt.Sprintf(":%s", strconv.Itoa(port))
	server.Run(addr)

}
