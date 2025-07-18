package api

import (
	"net/http"

	"github.com/bytedance/sonic"
)

func writeResponseMessage(w http.ResponseWriter, cod int, message string) {
	w.WriteHeader(cod)
	_ = sonic.ConfigFastest.NewEncoder(w).Encode(map[string]interface{}{
		"cod": cod,
		"msg": message,
	})
}

func writeErrorMessage(w http.ResponseWriter, cod int, err error) {
	w.WriteHeader(cod)
	_ = sonic.ConfigFastest.NewEncoder(w).Encode(map[string]interface{}{
		"cod":   cod,
		"error": err.Error(),
	})
}

// If there are no filter query params, return nil-map (compatible with repository)
func getFilterFromQuery(r *http.Request) map[string]interface{} {
	filter := make(map[string]interface{})
	if name := r.URL.Query().Get("name"); name != "" {
		filter["name"] = name
	}
	if uid := r.URL.Query().Get("uid"); uid != "" {
		filter["uid"] = uid
	}
	if len(filter) == 0 {
		return nil
	}
	return filter
}
