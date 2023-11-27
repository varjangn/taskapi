package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/varjangn/taskapi/api"
	"github.com/varjangn/taskapi/storage"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {

	store, err := storage.NewPostgresStorage()
	if err != nil {
		log.Fatal(err)
	}
	defer store.Db.Close()

	server := api.NewAPIServer(store)
	if err = server.Run(); err != nil {
		log.Fatal(err)
	}

}
