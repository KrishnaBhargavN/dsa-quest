// DSA Quest: a points-and-rewards tracker for DSA practice.
//
// One binary serves the API under /api and the built React app for every
// other path. The app is embedded from web/dist, so deploying needs only Go.
package main

import (
	"embed"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"os"
	"time"

	"dsaquest/internal/api"
	"dsaquest/internal/store"
)

//go:embed all:web/dist
var webDist embed.FS

func main() {
	_ = mime.AddExtensionType(".webmanifest", "application/manifest+json")

	static, err := fs.Sub(webDist, "web/dist")
	if err != nil {
		log.Fatal(err)
	}

	var st *store.Store
	dsn := firstSet("DATABASE_URL", "POSTGRES_URL")
	if dsn == "" {
		log.Print("warning: DATABASE_URL is not set; the API will answer 503 until it is")
	} else if st, err = store.Open(dsn); err != nil {
		log.Fatal(err)
	}

	password := os.Getenv("APP_PASSWORD")
	if password == "" {
		log.Print("warning: APP_PASSWORD is not set; nobody can log in until it is")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           api.New(st, password, static).Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Printf("DSA Quest listening on :%s", port)
	log.Fatal(srv.ListenAndServe())
}

func firstSet(keys ...string) string {
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return ""
}
