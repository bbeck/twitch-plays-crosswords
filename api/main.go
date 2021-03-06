package main

import (
	"github.com/bbeck/puzzles-with-chat/api/acrostic"
	"github.com/bbeck/puzzles-with-chat/api/crossword"
	"github.com/bbeck/puzzles-with-chat/api/pubsub"
	"github.com/bbeck/puzzles-with-chat/api/spellingbee"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gomodule/redigo/redis"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	pool := NewRedisPool()
	defer func() { _ = pool.Close() }()

	registry := new(pubsub.Registry)

	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Register handlers for our paths.
	r.Route("/api", func(r chi.Router) {
		RegisterRoutes(r, pool, registry)

		acrostic.RegisterRoutes(r, pool, registry)
		crossword.RegisterRoutes(r, pool, registry)
		spellingbee.RegisterRoutes(r, pool, registry)
	})

	// Start the server.
	err := http.ListenAndServe(":5000", r)
	if err != nil {
		log.Fatalf("error from main: %+v", err)
	}
}

func NewRedisPool() *redis.Pool {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = ":6379"
	}

	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 300 * time.Second,

		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", host)
		},

		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}
