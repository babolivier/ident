package main

import (
	"flag"
	"net/http"

	"github.com/babolivier/ident/common/config"
	"github.com/babolivier/ident/common/database"

	"github.com/sirupsen/logrus"
)

func main() {
	configFile := flag.String("config", "config.yaml", "Path to the configuration file")
	flag.Parse()

	// Load the configuration from the configuration file.
	cfg, err := config.NewConfig(*configFile)
	if err != nil {
		logrus.WithError(err).Fatal("Couldn't load the server configuration")
	}

	// Initiate the connection to the database and prepare statements.
	db, err := database.NewDatabase(cfg.Database.Driver, cfg.Database.ConnString)
	if err != nil {
		logrus.WithError(err).Fatal("Couldn't initiate a connection to the database")
	}

	router := NewRouter(cfg, db)

	logrus.WithField("listen_addr", cfg.HTTP.ListenAddr).Info("Starting up HTTP server")
	if err := http.ListenAndServe(cfg.HTTP.ListenAddr, router); err != nil {
		logrus.WithError(err).Fatal("Failed to serve http")
	}
}
