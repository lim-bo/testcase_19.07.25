package api

import (
	"net/http"
	"testcase/models"
	"time"

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

func getPeriodFromQuery(r *http.Request) (*models.RangeOpts, error) {
	var period models.RangeOpts
	var start, end string
	if start = r.URL.Query().Get("start"); start != "" {
		parsed, err := time.Parse("01-2006", start)
		if err != nil {
			return nil, err
		}
		period.Start = parsed
	}
	if end = r.URL.Query().Get("end"); end != "" {
		parsed, err := time.Parse("01-2006", end)
		if err != nil {
			return nil, err
		}
		period.End = parsed
	}
	if start == "" || end == "" {
		return nil, nil
	}
	return &period, nil
}
