package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/seggga/postgres/pkg/video-hint/service"
	"github.com/seggga/postgres/pkg/video-hint/storage"
)

func GetVideosByCaption(w http.ResponseWriter, r *http.Request, captionSubstring string) {
	dbIface := r.Context().Value(storage.ContextKeyDB)
	if dbIface == nil {
		log.Println("DB is not found in the request context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	db, ok := dbIface.(storage.DB)
	if !ok {
		log.Println("DB in the request context is not of type storage.DB")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	videos, err := service.GetVideosByCaption(db, captionSubstring)
	if err != nil {
		log.Println(err)
		if errors.Is(err, service.ErrIncorrectCaptionSubstring) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrDBRequestFailed) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp, err := json.Marshal(videos)
	if err != nil {
		log.Printf("failed to serialize videos list to JSON: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(resp); err != nil {
		log.Printf("failed to write videos list as a response body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
