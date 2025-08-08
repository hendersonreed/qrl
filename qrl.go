package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"time"
	qrcode "github.com/skip2/go-qrcode"
)

var logger *log.Logger

func initLogger() (*os.File, error) {
	logFile, err := os.OpenFile(".qrl.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	multi := io.MultiWriter(os.Stdout, logFile)
	logger = log.New(multi, "", log.LstdFlags|log.Lmicroseconds)
	return logFile, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	switch r.Method {
	case http.MethodGet:
		http.ServeFile(w, r, "./static/index.html")
		logger.Printf("GET / served index.html from %s in %.4fs", r.RemoteAddr, time.Since(start).Seconds())

	case http.MethodPut:
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			logger.Printf("PUT / failed to read body from %s: %v", r.RemoteAddr, err)
			return
		}
		defer r.Body.Close()

		png, err := qrcode.Encode(string(bodyBytes), qrcode.Medium, 256)
		if err != nil {
			http.Error(w, "Failed to generate QR code", http.StatusInternalServerError)
			logger.Printf("PUT / failed to generate QR code from %s: %v", r.RemoteAddr, err)
			return
		}

		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(http.StatusOK)
		w.Write(png)
		logger.Printf("PUT / generated QR code from %s in %.4fs", r.RemoteAddr, time.Since(start).Seconds())

	default:
		http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
		logger.Printf("Unsupported method %s on / from %s", r.Method, r.RemoteAddr)
	}
}

func main() {
	logFile, err := initLogger()
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	http.HandleFunc("/", handler)
	logger.Println("Server starting on :8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		logger.Fatalf("Server error: %v", err)
	}
}
