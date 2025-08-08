package main

import (
    "fmt"
    "net/http"
    "log"
    qrcode "github.com/skip2/go-qrcode"
)


func handler(w http.ResponseWriter, r *http.Request) {
    if (r.Method != "PUT") {
        fmt.Fprintf(w, "This is qrl - curl qrl to generate a QR code")
    }
    else {
        var png []byte
        png, err := qrcode.Encode(r.Body, qrcode.Medium, 256)
    }
}

func main() {
    http.HandleFunc("/", handler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
