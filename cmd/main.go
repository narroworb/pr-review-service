package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/narroworb/pr-review-service/internal/database"
	"github.com/narroworb/pr-review-service/internal/handlers"
	"github.com/narroworb/pr-review-service/internal/middleware"
)

func main() {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		log.Fatal("empty environment variable POSTGRES_DSN")
	}
	db, err := database.NewPostgresDB(dsn)
	if err != nil {
		log.Fatalf("error in creation db: %v", err)
	}
	defer db.Close()
	log.Println("Connection to Postgres established")
	err = db.RunMigrations()
	if err != nil {
		log.Fatalf("error in migrations: %v", err)
	}
	log.Println("Migrations to Postgres applied")

	h := handlers.NewHandlersRepo(db)

	r := chi.NewRouter()

	r.Use(middleware.TimeoutMiddleware(3 * time.Second))

	r.Post("/team/add", h.AddTeam)
	r.Get("/team/get", h.GetTeam)

	r.Post("/users/setIsActive", h.SetUserIsActive)
	r.Get("/users/getReview", h.GetReview)

	r.Post("/pullRequest/create", h.CreatePR)
	r.Post("/pullRequest/merge", h.MergePR)
	r.Post("/pullRequest/reassign", h.ReassignPR)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Println("listening on :8080")
		if err := http.ListenAndServe(":8080", r); err != nil && err != http.ErrServerClosed {
			log.Fatalf("error in server work: %v", err)
		}
	}()

	<-stop

	log.Println("shutting down: stopping to accept new requests...")
}
