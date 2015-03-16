package server

import (
	"bartlett/shared"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type State struct {
	Files map[string]shared.File // keys are directory path
}

func statictest(w http.ResponseWriter, r *http.Request) {
	js, err := json.Marshal(static)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func livetest(w http.ResponseWriter, r *http.Request) {
	js, err := json.Marshal(live)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func files(w http.ResponseWriter, r *http.Request) {
	resp := make([]string, 0)

	for key, _ := range static.Files {
		resp = append(resp, key)
	}

	js, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func Run(wg *sync.WaitGroup, port int) {
	defer wg.Done()
	log.Println("starting server on port", port)

	// static handlers
	static = State{Files: make(map[string]shared.File)}
	http.HandleFunc("/sync", Sync)
	http.HandleFunc("/read", Read)

	// live handlers
	live = State{Files: make(map[string]shared.File)}
	http.HandleFunc("/watch", Watch)
	http.HandleFunc("/push", Push)
	http.HandleFunc("/unlock", Unlock)

	// testing handlers
	http.HandleFunc("/static", statictest)
	http.HandleFunc("/live", livetest)
	http.HandleFunc("/files", files)

	// listen and serve
	portString := ":" + strconv.FormatInt(int64(port), 10)
	log.Println("listen and serve on", portString)
	err := http.ListenAndServe(portString, nil)
	if err != nil {
		log.Fatal(err)
	}
}
