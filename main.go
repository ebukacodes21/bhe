package main

import (
	"bhe/api"
	db "bhe/db/sqlc"
	"bhe/helper"
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	config, err := helper.LoadConfig(".")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal(err)
	}

	repository := db.NewRepository(conn)
	server, err := api.NewServer(config, repository)
	if err != nil {
		log.Fatal("cannot create server", err)
	}

	err = server.Start(config.Addr)
	if err != nil {
		log.Fatal("cannot start server", err)
	}
}
