// PicknShip — mobile-first pick-and-ship demo for Wesco Seeds.
package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/wescoseeds/picknship/internal/handlers"
	"github.com/wescoseeds/picknship/internal/store"
)

//go:embed templates/*.html
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

func main() {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}

	st := store.New()

	srv, err := handlers.New(st, templatesFS)
	if err != nil {
		log.Fatalf("templates: %v", err)
	}

	mux := http.NewServeMux()
	srv.Routes(mux)

	staticSub, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatalf("static fs: %v", err)
	}
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))

	log.Printf("PicknShip listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
