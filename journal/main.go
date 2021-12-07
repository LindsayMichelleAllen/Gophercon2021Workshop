package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

var (
	//go:embed html/last.html
	lastHTML     string
	lastTemplate = template.Must(template.New("last").Parse(lastHTML))
	apiKey       string
)

type Server struct {
	db *DB
}

func (s *Server) lastHandler(w http.ResponseWriter, r *http.Request) {

	entry, err := s.db.Last(r.Context())
	if err != nil {
		fmt.Fprintf(w, "No entries")
		return
	}

	if err := lastTemplate.Execute(w, entry); err != nil {
		log.Printf("template: %s", err)
	}
	/*
		time := entry.Time.Format("2006-01-02T15:04")
		content := html.EscapeString(entry.Content)
		lastText = fmt.Sprintf("[%s] %s by %s", time, content, entry.User)
		fmt.Fprintf(w, lastHTML, lastText)
	*/
}

// TODO: Protect this handler with Bearer authentication
// curl -H 'Authorization: Bearer s3cr3t' -d@add-1.json http://localhost:8080/new
// Extra: Write a middleware
func (s *Server) newHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	hdr := r.Header.Get("Authorization")
	key := strings.TrimPrefix(hdr, "Bearer ")
	if key != apiKey {
		http.Error(w, "bad auth", http.StatusUnauthorized)
		return
	}

	var e Entry

	const maxSize = 1 << 20 // 1MB
	rdr := io.LimitReader(r.Body, maxSize)
	if err := json.NewDecoder(rdr).Decode(&e); err != nil {
		// Rule of thumb: log error, return generic messages
		// TODO: Request ID
		log.Printf("decode json: %s", err)
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	if err := e.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	e.Time = time.Now()
	if err := s.db.Add(r.Context(), e); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Other option: Use json.Marshal, check error and then w.Write
	if err := json.NewEncoder(w).Encode(e); err != nil {
		log.Printf("encode: %s", err)
	}
}

func (s *Server) queryHandler(w http.ResponseWriter, r *http.Request) {
	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")
	if start == "" || end == "" {
		http.Error(w, "missing start or end", http.StatusBadRequest)
		return
	}

	startTime, err := time.Parse("2006-01-02", start)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	endTime, err := time.Parse("2006-01-02", end)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	entries, err := s.db.Query(r.Context(), startTime, endTime)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(entries); err != nil {
		log.Printf("encode: %s", err)
	}
}

func (s *Server) Health(ctx context.Context) error {
	return s.db.Health(ctx)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	if err := s.Health(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "OK\n")
}

func main() {
	apiKey = os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("missing API_KEY")
	}
	var err error
	db, err := NewDB("user=postgres password=s3cr3t sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	s := Server{db}

	r := mux.NewRouter()
	r.HandleFunc("/last", s.lastHandler).Methods("GET")
	r.HandleFunc("/new", s.newHandler).Methods("POST")
	r.HandleFunc("/query", s.queryHandler).Methods("GET")
	r.HandleFunc("/health", s.healthHandler).Methods("GET")

	http.Handle("/", r)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,

		ReadTimeout:       2 * time.Second,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 1 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
