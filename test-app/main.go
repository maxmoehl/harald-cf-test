package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8080"
	}

	err := http.ListenAndServe(":"+port, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`"ok"`))
		if err != nil {
			fmt.Println("unable to write response:", err.Error())
		}
	}))

	if err != nil {
		fmt.Println("error:", err.Error())
		os.Exit(1)
	}
}
