package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Coaster struct {
	Name         string `json:"name,omitempty"`
	Manufacturer string `json:"manufacturer,omitempty"`
	ID           string `json:"id,omitempty"`
	Height       int    `json:"height,omitempty"`
}

type coasterHandlers struct {
	sync.Mutex
	store map[string]Coaster
}

func (h *coasterHandlers) coasters(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		h.get(w, r)
		return
	case "POST":
		h.post(w, r)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method forbidden"))
		return

	}
}

//Get function that returns all coasters in store
func (h *coasterHandlers) get(w http.ResponseWriter, r *http.Request) {
	coasters := make([]Coaster, len(h.store))

	i := 0
	h.Lock()
	for _, coaster := range h.store {
		coasters[i] = coaster
		i++
	}
	h.Unlock()

	jsonBytes, err := json.Marshal(coasters)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(([]byte(err.Error())))
	}
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *coasterHandlers) getCoaster(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.String(), "/")

	if len(parts) != 3 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	h.Lock()
	coaster, ok := h.store[parts[2]]

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	h.Unlock()

	jsonBytes, err := json.Marshal(coaster)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(([]byte(err.Error())))
	}
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)

}

//Get function that returns all coasters in store
func (h *coasterHandlers) post(w http.ResponseWriter, r *http.Request) {
	//Bodybytes just contains json user provides
	bodyBytes, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	ct := r.Header.Get("content-type")
	if ct != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		w.Write([]byte("Requires json input"))
		return
	}

	var coaster Coaster
	err = json.Unmarshal(bodyBytes, &coaster)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	coaster.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	h.Lock()
	h.store[coaster.ID] = coaster
	h.Unlock()

}

func newCoasterHandlers() *coasterHandlers {
	return &coasterHandlers{
		store: map[string]Coaster{},
	}
}

func main() {
	coasterHandlers := newCoasterHandlers()
	http.HandleFunc("/coasters", coasterHandlers.coasters)
	http.HandleFunc("/coasters/", coasterHandlers.getCoaster)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
