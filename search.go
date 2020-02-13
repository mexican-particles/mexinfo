package mexinfo

import (
	"encoding/json"
	"log"
	"net/http"
)

// MexSearch brings Mexican information to the Sumo world
func MexSearch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode("hoge"); err != nil {
		log.Fatalf("json.Marshal: %v", err)
	}
}
