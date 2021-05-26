package templates

var Main = `package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/99designs/gqlgen/handler"
	"github.com/rs/cors"
)

func main() {
	app := cli.NewApp()
	app.Name = "AYaro Thesis"
	app.Usage = "Generate server app"

	app.Commands = []cli.Command{
		startCmd,
		automigrateCmd,
	}

	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}

var startCmd = cli.Command{
	Name:  "start",
	Usage: "start http server",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:   "p,port",
			Usage:  "Port to listen to",
			Value:  "80",
			EnvVar: "PORT",
		},
	},

	Action: func(ctx *cli.Context) error {
		port := ctx.String("port")
		if err := startHttpServer(port); err != nil {
			return cli.NewExitError(err.Error(), 1)
		}

		return nil
	},
}

var automigrateCmd = cli.Command{
	Name:  "automigrate",
	Usage: "gorm automigration",
	Action: func(ctx *cli.Context) error {

		db := gen.NewDBFromEnvVars()
		defer db.Close()
		return db.AutoMigrate()

		if err := automigrate(); err != nil {
			return cli.NewExitError(err.Error(), 1)
		}
		return nil
	},
}


func startHttpServer(port string) error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	db := gen.NewDBFromEnvVars()
	defer db.Close()

	eventController, err := gen.NewEventController()
	if err != nil {
		return err
	}

	mux := gen.GetHTTPServeMux(src.New(db, &eventController), db, src.GetMigrations(db))
	handler := cors.AllowAll().Handler(mux)
	
	h := &http.Server{Addr: ":" + port, Handler: handler}

	go func() {
		log.Printf("connect to http://localhost:%s/graphql for GraphQL playground", port)
		log.Fatal(h.ListenAndServe())
	}()

	<-stop

	log.Println("\n Shutting down the server...")

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	err = h.Shutdown(ctx)
	if err != nil {
		return cli.NewExitError(err, 1)
	}

	log.Println("Server gracefully stopped")

	err = db.Close()
	if err != nil {
		return cli.NewExitError(err, 1)
	}
	log.Println("Database connection closed")

	return nil
}
`
