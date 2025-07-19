// @title Subs-API
// @version 1.0
// @description API-service for managing users' subscriptions
// @schemes http
package main

import (
	"log"
	"testcase/internal/api"
	"testcase/internal/settings"
	"testcase/internal/subs"
)

func main() {
	cfg := settings.GetConfig()
	sr := subs.New(&subs.DBConfig{
		Address:  cfg.GetString("db_addr"),
		User:     cfg.GetString("db_user"),
		Password: cfg.GetString("db_pass"),
		DBName:   cfg.GetString("db_name"),
	})

	serv := api.New(sr)
	log.Fatal(serv.Run(cfg.GetString("api_address")))
}
