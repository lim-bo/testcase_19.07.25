// @title Subs-API
// @version 1.0
// @description API-service for managing users' subscriptions
// @schemes http
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"testcase/internal/api"
	"testcase/internal/settings"
	"testcase/internal/subs"
	"time"
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
	servError := make(chan error, 1)
	go func() {
		if err := serv.Run(cfg.GetString("api_address")); err != nil && err != http.ErrServerClosed {
			servError <- err
		}
	}()

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	select {
	case <-stopChan:
		log.Println("Shutdown signal recieved")
	case err := <-servError:
		log.Println("Server error: ", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	log.Println("Shutting down..")
	if err := serv.Shutdown(ctx); err != nil {
		log.Println("Shutdown failed with error: " + err.Error())
	} else {
		log.Println("Server stopped")
	}
}
