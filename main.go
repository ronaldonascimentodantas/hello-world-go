package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			fmt.Printf("%s %s 404\n", r.Method, r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodGet {
			fmt.Printf("%s %s 405\n", r.Method, r.URL.Path)
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		fmt.Printf("%s %s 200\n", r.Method, r.URL.Path)
		fmt.Fprint(w, "Hello World")
	})
	http.ListenAndServe(":"+port, nil)
}
