package web

import (
	"encoding/json"
	"net/http"
	"time"
)

type statusInfo struct {
	StartupTime time.Time
	Info        map[string]interface{}
}

func (s *Server) infoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(s.info)
}
