package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!")
}

func calcHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	aStr := q.Get("a")
	bStr := q.Get("b")
	op := q.Get("op")

	a, err := strconv.ParseFloat(aStr, 64)
	if err != nil {
		http.Error(w, "invalid parameter: a", http.StatusBadRequest)
		return
	}
	b, err := strconv.ParseFloat(bStr, 64)
	if err != nil {
		http.Error(w, "invalid parameter: b", http.StatusBadRequest)
		return
	}

	var result float64
	switch op {
	case "add":
		result = a + b
	case "sub":
		result = a - b
	case "mul":
		result = a * b
	case "div":
		if b == 0 {
			http.Error(w, "division by zero", http.StatusBadRequest)
			return
		}
		result = a / b
	default:
		http.Error(w, "invalid op: use add, sub, mul, div", http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "%g", result)
}

func main() {
	http.HandleFunc("/", helloHandler)
	http.HandleFunc("/calc", calcHandler)

	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}