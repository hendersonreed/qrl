package main

import (
	"io"
	"log"
	"net/http"
	"strings"
	"unicode/utf8"
	"errors"
	"time"

	qrcode "github.com/skip2/go-qrcode"
)

func sanitizeInput(input string) (string, error) {
    if !utf8.ValidString(input) {
        return "", errors.New("invalid UTF-8 input")
    }

    input = strings.TrimSpace(input)
    if len(input) == 0 {
        return "", errors.New("empty input not allowed")
    }

    for _, r := range input {
        if (r < 32 || r == 127) && r != '\n' && r != '\r' && r != '\t' {
            return "", errors.New("input contains invalid control characters")
        }
    }

    return input, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	switch r.Method {
	case http.MethodGet:
		http.ServeFile(w, r, "./web/static/index.html")
		log.Printf("GET / served index.html from %s in %.4fs", r.RemoteAddr, time.Since(start).Seconds())

	case http.MethodPut:
		r.Body = http.MaxBytesReader(w, r.Body, 3000)
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
			log.Printf("PUT / request body too large from %s: %v", r.RemoteAddr, err)
			return
		}
		defer r.Body.Close()

		input, err := sanitizeInput(string(bodyBytes))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			log.Printf("PUT / invalid input from %s: %v", r.RemoteAddr, err)
			return
		}

		png, err := qrcode.Encode(input, qrcode.Medium, 256)
		if err != nil {
			http.Error(w, "Failed to generate QR code", http.StatusInternalServerError)
			log.Printf("PUT / failed to generate QR code from %s: %v", r.RemoteAddr, err)
			return
		}

		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(http.StatusOK)
		w.Write(png)
		log.Printf("PUT / generated QR code from %s", r.RemoteAddr)

	default:
		http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
		log.Printf("Unsupported method %s on / from %s", r.Method, r.RemoteAddr)
	}
}

func main() {
	http.HandleFunc("/", handler)
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
