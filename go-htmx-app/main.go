package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	gRouter := mux.NewRouter()
	gRouter.HandleFunc("/", HomeHandler).Methods("GET")

	fmt.Println("Server is running on port 3000")
	http.ListenAndServe(":3000", gRouter)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the home page")
}
