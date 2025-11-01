package main

import (
	"net/http"
)

func (cfg *apiConfig) handlerListChirps(w http.ResponseWriter, r *http.Request) {
	dbChirps, err := cfg.db.ListChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching chirps", err)
		return
	}

	chirps := []*Chirp{}
	for _, chirp := range dbChirps {
		chirps = append(chirps, fromDbChirp(&chirp))
	}

	respondWithJSON(w, http.StatusOK, &chirps)
}
