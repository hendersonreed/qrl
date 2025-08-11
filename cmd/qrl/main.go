package main

import (
	"io"
	"log"
	"net/http"
	"time"

	qrcode "github.com/skip2/go-qrcode"
)

func handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	switch r.Method {
	case http.MethodGet:
		http.ServeFile(w, r, "./static/index.html")
		log.Printf("GET / served index.html from %s in %.4fs", r.RemoteAddr, time.Since(start).Seconds())

	case http.MethodPut:
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			log.Printf("PUT / failed to read body from %s: %v", r.RemoteAddr, err)
			return
		}
		defer r.Body.Close()

		png, err := qrcode.Encode(string(bodyBytes), qrcode.Medium, 256)
		if err != nil {
			http.Error(w, "Failed to generate QR code", http.StatusInternalServerError)
			log.Printf("PUT / failed to generate QR code from %s: %v", r.RemoteAddr, err)
			return
		}

		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(http.StatusOK)
		w.Write(png)
		log.Printf("PUT / generated QR code from %s in %.4fs", r.RemoteAddr, time.Since(start).Seconds())

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
