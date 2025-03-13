package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var rdb *redis.Client

const shortURLPrefix = "http://localhost:8080/"

// Generate a random short code
func generateShortCode() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, 6)
	for i := range code {
		code[i] = charset[rand.Intn(len(charset))]
	}
	return string(code)
}

// Shorten URL Handler
func shortenURL(w http.ResponseWriter, r *http.Request) {
	longURL := r.URL.Query().Get("url")
	if longURL == "" {
		http.Error(w, "Missing URL", http.StatusBadRequest)
		return
	}

	shortCode := generateShortCode()
	rdb.Set(ctx, shortCode, longURL, 0)

	fmt.Fprintf(w, "Shortened URL: %s%s", shortURLPrefix, shortCode)
}

// Expand URL Handler
func expandURL(w http.ResponseWriter, r *http.Request) {
	shortCode := r.URL.Path[1:]
	longURL, err := rdb.Get(ctx, shortCode).Result()
	if err == redis.Nil {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, longURL, http.StatusFound)
}

func main() {
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	http.HandleFunc("/shorten", shortenURL)
	http.HandleFunc("/", expandURL)

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
