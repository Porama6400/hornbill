package main

import (
	"context"
	"github.com/joho/godotenv"
	"hornbill/pkg/apiserver"
	"log"
	"os"
	"strings"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println(".env file not loaded")
	}

	daemonList := strings.Split(os.Getenv("DAEMON_LIST"), ";")
	server, err := apiserver.NewServer(daemonList)
	if err != nil {
		log.Fatal(err)
	}

	err = server.PingAll(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Successfully connected to daemon")

}
