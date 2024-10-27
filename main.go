package main

import (
	"fmt"
	"github.com/Scalingo/go-handlers"
	"github.com/Scalingo/go-utils/logger"
	"github.com/Scalingo/sclng-backend-test-v1/handle"
	"github.com/joho/godotenv"
	"net/http"
	"os"
)

func main() {
	log := logger.Default()
	log.Info("Initializing app")
	cfg, err := newConfig()
	if err != nil {
		log.WithError(err).Error("Fail to initialize configuration")
		os.Exit(1)
	}

	// load .env file
	err = godotenv.Load(".env")
	if err != nil {
		log.WithError(err).Error("Fail to load .env file")
	}

	log.Info("Initializing routes")
	router := handlers.NewRouter(log)
	//router.Use(handlers.MiddlewareFunc(middleware.CorrelationMiddleware))
	router.HandleFunc("/ping", handle.PongHandler)
	// initialize ReposHandler implementing the Handler interface
	reposHandler, err := handle.InitReposHandler(log, 100)
	if err != nil {
		log.WithError(err).Error("Fail to config repos handler")
		os.Exit(1)
	}
	router.Handle("/repos", reposHandler)

	log = log.WithField("port", cfg.Port)
	log.Info("Listening...")
	err = http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), router)
	if err != nil {
		log.WithError(err).Error("Fail to listen to the given port")
		os.Exit(2)
	}
}
