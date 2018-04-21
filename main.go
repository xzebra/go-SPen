package main

import (
	"net/http"
)

func handlerFavicon(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "public/favicon.ico")
}

func handler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "public/index.html")
}

func main() {
	http.HandleFunc("/favicon.ico", handlerFavicon)
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
