package main

import (
	"api/cmd/helper"
	"api/cmd/router"
	"api/database"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	//initialize database connection
	if err := database.Init(); err != nil {
		log.Fatal(err)
	}

	defer database.Close()

	issuer, err := helper.NewIssuer(os.Getenv("JWT_CERT_PATH"))
	if err != nil {
		fmt.Printf("unable to create issuer: %v\n", err)
		log.Fatal(err)
	}

	server := router.New(database.DB, issuer)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := server.Serve(ctx); err != nil {
		log.Fatal(err)
	}
}
