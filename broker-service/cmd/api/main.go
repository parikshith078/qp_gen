package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/parikshith078/qp_gen/broker/internal/db"
	"github.com/parikshith078/qp_gen/broker/internal/db/sqlc"
)

var webPort = os.Getenv("WEB_PORT")

type Config struct {
	Db *sqlc.Queries
}

var dsn = os.Getenv("DSN")

func main() {
	// run migrations
	db.RunMigrations(dsn)

	ctx := context.Background()
	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		log.Fatal("DB connection failed")
		panic(err)
	}
	defer conn.Close(ctx)
	repository := sqlc.New(conn)
	app := Config{
		Db: repository,
	}

	log.Printf("Starting broker service on port %s\n", webPort)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err = srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
