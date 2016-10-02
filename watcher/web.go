package main

import (
	"fmt"
	"net/http"
	"strings"
)

func randomSentinelHandler(w http.ResponseWriter, r *http.Request) {
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

func allSentinelsHandler(w http.ResponseWriter, r *http.Request) {
	keeper := Keeper{GetZookeeperHosts()}
	as, err := keeper.GetAllSentinels()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, err.Error())
		return
	}
	s := make([]string, len(as))
	for i, sent := range as {
		s[i] = sent.GetDottedAddr()
	}
	fmt.Fprintf(w, strings.Join(s, ","))
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func serve() {
	http.HandleFunc("/sentinel", randomSentinelHandler)
	http.HandleFunc("/sentinels", allSentinelsHandler)
	http.HandleFunc("/", statusHandler)
	http.ListenAndServe(":8080", nil)
}
