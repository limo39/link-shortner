package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type ShortURL struct {
	ID       string `json:"id"`
	Original string `json:"original"`
	Short    string `json:"short"`
}

type ShortenRequest struct {
	URL string `json:"url"`
}

var urlStore = make(map[string]string)
var baseURL = "http://localhost:8080/"

func main() {
	rand.Seed(time.Now().UnixNano())

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/shorten", shortenHandler)
	http.HandleFunc("/shorten-form", shortenFormHandler)
	http.HandleFunc("/s/", redirectHandler)

	fmt.Println("URL Shortener running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	tmpl := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>URL Shortener</title>
		<style>
			body { font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; }
			h1 { color: #333; }
			form { display: flex; flex-direction: column; gap: 10px; }
			input { padding: 8px; font-size: 16px; }
			button { padding: 10px; background: #0066cc; color: white; border: none; cursor: pointer; }
			.result { margin-top: 20px; padding: 10px; background: #f0f0f0; }
		</style>
	</head>
	<body>
		<h1>URL Shortener</h1>
		<form action="/shorten-form" method="post">
			<input type="url" name="url" placeholder="Enter URL to shorten" required>
			<button type="submit">Shorten URL</button>
		</form>
	</body>
	</html>
	`

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, tmpl)
}

func shortenFormHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	originalURL := r.FormValue("url")
	if originalURL == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	shortID := generateShortID(6)
	urlStore[shortID] = originalURL

	shortURL := baseURL + "s/" + shortID

	tmpl := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>URL Shortened</title>
		<style>
			body { font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; }
			.result { margin-top: 20px; padding: 20px; background: #f0f0f0; }
			a { color: #0066cc; word-break: break-all; }
		</style>
	</head>
	<body>
		<h1>URL Shortened</h1>
		<div class="result">
			<p><strong>Original URL:</strong><br>%s</p>
			<p><strong>Short URL:</strong><br><a href="%s">%s</a></p>
		</div>
		<p><a href="/">Shorten another URL</a></p>
	</body>
	</html>
	`

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, tmpl, originalURL, shortURL, shortURL)
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	shortID := generateShortID(6)
	urlStore[shortID] = req.URL

	resp := ShortURL{
		ID:       shortID,
		Original: req.URL,
		Short:    baseURL + "s/" + shortID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	shortID := r.URL.Path[len("/s/"):]
	if shortID == "" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	originalURL, exists := urlStore[shortID]
	if !exists {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusMovedPermanently)
}

func generateShortID(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
