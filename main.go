package main

import (
	"os"

	"github.com/Coderx44/MovieTicketingPortal/app"
	"github.com/Coderx44/MovieTicketingPortal/config"
	"github.com/Coderx44/MovieTicketingPortal/db"
	"github.com/Coderx44/MovieTicketingPortal/service"
	logger "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func main() {
	logger.SetFormatter(&logger.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "02-01-2006 15:04:05",
	})

	config.Load()
	app.Init()
	defer app.Close()

	cliApp := cli.NewApp()
	cliApp.Name = config.AppName()
	cliApp.Version = "1.0.0"
	cliApp.Commands = []cli.Command{
		{
			Name:  "start",
			Usage: "start server",
			Action: func(c *cli.Context) {
				service.StartApiServer()
			},
		},
		{
			Name:  "create_migration",
			Usage: "create migration file",
			Action: func(c *cli.Context) error {
				return db.CreateMigrationFile(c.Args().Get(0))
			},
		},
		{
			Name:  "migrate",
			Usage: "run db migrations",
			Action: func(c *cli.Context) error {
				return db.RunMigrations()
			},
		},
		{
			Name:  "rollback",
			Usage: "rollback migrations",
			Action: func(c *cli.Context) error {
				return db.RollbackMigration(c.Args().Get(0))
			},
		},
	}

	if err := cliApp.Run(os.Args); err != nil {
		panic(err)
	}
}
