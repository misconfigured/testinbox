package main

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
)

var db *pgxpool.Pool

func connectDB() {
	var err error
	db, err = pgxpool.Connect(context.Background(), "postgres://testinbox@host.docker.internal:5432/testinbox_development")
	if err != nil {
		log.WithFields(logrus.Fields{
			"module": "database",
			"action": "connectDB",
			"error":  err,
		}).Fatal("Unable to connect to database")
	} else {
		log.WithFields(logrus.Fields{
			"module": "database",
			"action": "connectDB",
		}).Info("Database connection established successfully")
	}
}
