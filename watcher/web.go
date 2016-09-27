package main

import (
	"fmt"
	"net/http"
)

func sentinelHandler(w http.ResponseWriter, r *http.Request) {
	keeper := Keeper{GetZookeeperHosts()}
	nr, err := keeper.GetRandomSentinel()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, err.Error())
		return
	}
	// REVIEW: GetAddress? GetDottedAddr? Serialize into JSON?
	fmt.Fprintf(w, nr.GetAddress())
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hi")
}

func serve() {
	http.HandleFunc("/sentinel", sentinelHandler)
	http.HandleFunc("/", statusHandler)
	http.ListenAndServe(":8080", nil)
}
