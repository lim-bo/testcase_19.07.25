package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"testcase/internal/errvalues"
	"testcase/models"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
)

func (s *Server) CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, GET, POST, DELETE, PUT")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := uuid.New()
		ctx := context.WithValue(r.Context(), "Request-ID", reqID.String())
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func (s *Server) subIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Context().Value("Request-ID").(string)
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			slog.Error("incoming request with invalid id",
				slog.String("req_id", reqID),
				slog.String("from", r.RemoteAddr))
			writeErrorMessage(w, http.StatusBadRequest, errvalues.ErrInvalidRequest)
			return
		}
		ctx := context.WithValue(r.Context(), "Sub-ID", id)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

// @Summary Registering subscription
// @Description Recieves new subscription info and
// @Description saves it in DB
// @Tags subs
// @Router /subs/add [post]
// @Accept json
// @Produce json
// @Param request body models.Subscription true "New subscription data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
func (s *Server) addSubscription(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	reqID := r.Context().Value("Request-ID").(string)
	var sub models.Subscription
	err := sonic.ConfigDefault.NewDecoder(r.Body).Decode(&sub)
	if err != nil {
		slog.Error("error decoding request body",
			slog.String("error", err.Error()),
			slog.String("req_id", reqID),
			slog.String("from", r.RemoteAddr))
		writeErrorMessage(w, http.StatusBadRequest, errvalues.ErrInvalidRequest)
		return
	}
	err = s.subsRepo.AddSub(&sub)
	if err != nil {
		slog.Error("error adding subscription",
			slog.String("error", err.Error()),
			slog.String("req_id", reqID),
			slog.String("from", r.RemoteAddr))
		writeErrorMessage(w, http.StatusInternalServerError, errvalues.ErrInternal)
		return
	}
	slog.Info("successfully added new subscription",
		slog.String("req_id", reqID),
		slog.String("from", r.RemoteAddr))
	writeResponseMessage(w, http.StatusOK, "sub added")
}

// @Summary Getting subcription info
// @Description Provides subscription data by ID in path
// @Tags subs
// @Router /subs/{id} [get]
// @Param id path int true "Subscription ID"
// @Produce json
// @Success 200 {object} models.Subscription
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
func (s *Server) getSubscription(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	reqID := r.Context().Value("Request-ID").(string)
	subID := r.Context().Value("Sub-ID").(int)
	sub, err := s.subsRepo.GetSub(subID)
	if err != nil {
		if errors.Is(err, errvalues.ErrNoSuchRow) {
			slog.Error("get sub request with unexisted id",
				slog.String("req_id", reqID),
				slog.String("from", r.RemoteAddr))
			writeErrorMessage(w, http.StatusBadRequest, err)
			return
		}
		slog.Error("error getting subscription",
			slog.String("error", err.Error()),
			slog.String("req_id", reqID),
			slog.String("from", r.RemoteAddr))
		writeErrorMessage(w, http.StatusInternalServerError, errvalues.ErrInternal)
		return
	}
	err = sonic.ConfigDefault.NewEncoder(w).Encode(sub)
	if err != nil {
		slog.Error("error providing result",
			slog.String("error", err.Error()),
			slog.String("req_id", reqID),
			slog.String("from", r.RemoteAddr))
		writeErrorMessage(w, http.StatusInternalServerError, errvalues.ErrInternal)
		return
	}
	slog.Info("successfully provided subscription info",
		slog.String("req_id", reqID),
		slog.String("from", r.RemoteAddr))
}

// @Summary Updating subscription
// @Description Recieves new subscription info for update
// @Description by provided id in path
// @Tags subs
// @Router /subs/{id} [put]
// @Accept json
// @Produce json
// @Param request body models.Subscription true "New subscription data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
func (s *Server) updateSubscription(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	reqID := r.Context().Value("Request-ID").(string)
	subID := r.Context().Value("Sub-ID").(int)
	var sub models.Subscription
	err := sonic.ConfigDefault.NewDecoder(r.Body).Decode(&sub)
	if err != nil {
		slog.Error("error decoding request body",
			slog.String("error", err.Error()),
			slog.String("req_id", reqID),
			slog.String("from", r.RemoteAddr))
		writeErrorMessage(w, http.StatusBadRequest, errvalues.ErrInvalidRequest)
		return
	}
	err = s.subsRepo.UpdateSub(subID, &sub)
	if err != nil {
		if errors.Is(err, errvalues.ErrNoSuchRow) {
			slog.Error("update sub request with unexisted id",
				slog.String("req_id", reqID),
				slog.String("from", r.RemoteAddr))
			writeErrorMessage(w, http.StatusBadRequest, err)
			return
		}
		slog.Error("error updating subscription",
			slog.String("error", err.Error()),
			slog.String("req_id", reqID),
			slog.String("from", r.RemoteAddr))
		writeErrorMessage(w, http.StatusInternalServerError, errvalues.ErrInternal)
		return
	}
	slog.Info("subscription successfully updated",
		slog.String("req_id", reqID),
		slog.String("from", r.RemoteAddr))
	writeResponseMessage(w, http.StatusOK, "subscription updated")
}

// @Summary Deleting subcription
// @Description Deletes subscription with given id
// @Tags subs
// @Router /subs/{id} [delete]
// @Param id path int true "Subscription ID"
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
func (s *Server) deleteSubscription(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	reqID := r.Context().Value("Request-ID").(string)
	subID := r.Context().Value("Sub-ID").(int)
	err := s.subsRepo.DeleteSub(subID)
	if err != nil {
		if errors.Is(err, errvalues.ErrNoSuchRow) {
			slog.Error("delete sub request with unexisted id",
				slog.String("req_id", reqID),
				slog.String("from", r.RemoteAddr))
			writeErrorMessage(w, http.StatusBadRequest, err)
			return
		}
		slog.Error("error deleting subscription",
			slog.String("error", err.Error()),
			slog.String("req_id", reqID),
			slog.String("from", r.RemoteAddr))
		writeErrorMessage(w, http.StatusInternalServerError, errvalues.ErrInternal)
		return
	}
	slog.Info("subscription successfully deleted",
		slog.String("req_id", reqID),
		slog.String("from", r.RemoteAddr))
	writeResponseMessage(w, http.StatusOK, "subscription deleted")
}

// @Summary Listing subscriptions
// @Description Returns list of subscriptions with given
// @Description unnecessary filters
// @Tags subs
// @Router /subs/list [get]
// @Param name query string false "Sub's service name" Example(Spotify)
// @Param uid query string false "User ID" Example(60601fee-2bf1-4721-ae6f-7636e79a0cba)
// @Param limit query int false "Returned rows limit"
// @Param offset query int false "Offset for paginations"
// @Param order query string false "Filed name for sorting by" Example(name, uid, id, created_at, expires, price)
// @Produce json
// @Success 200 {array} models.Subscription
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
func (s *Server) listSubscriptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	reqID := r.Context().Value("Request-ID").(string)
	filter := getFilterFromQuery(r)
	var limit, offset int
	var err error
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			slog.Error("incoming request with invalid query param",
				slog.String("req_id", reqID),
				slog.String("from", r.RemoteAddr))
			writeErrorMessage(w, http.StatusBadRequest, errvalues.ErrInvalidRequest)
			return
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			slog.Error("incoming request with invalid query param",
				slog.String("req_id", reqID),
				slog.String("from", r.RemoteAddr))
			writeErrorMessage(w, http.StatusBadRequest, errvalues.ErrInvalidRequest)
			return
		}
	}
	list, err := s.subsRepo.ListSubs(&models.ListOpts{
		Limit:  limit,
		Offset: offset,
		Filter: filter,
		Order:  r.URL.Query().Get("order"),
	})
	if err != nil {
		slog.Error("list subscriptions error",
			slog.String("error", err.Error()),
			slog.String("req_id", reqID),
			slog.String("from", r.RemoteAddr))
		writeErrorMessage(w, http.StatusInternalServerError, errvalues.ErrInternal)
		return
	}
	err = sonic.ConfigDefault.NewEncoder(w).Encode(list)
	if err != nil {
		slog.Error("error providing result",
			slog.String("error", err.Error()),
			slog.String("req_id", reqID),
			slog.String("from", r.RemoteAddr))
		writeErrorMessage(w, http.StatusInternalServerError, errvalues.ErrInternal)
		return
	}
	slog.Info("successfully listed subscriptions",
		slog.String("req_id", reqID),
		slog.String("from", r.RemoteAddr))
}

type sumResponse struct {
	Sum int `json:"sum" example:"1000"`
}

// @Summary Getting price sum
// @Description Recieving summary subscriptions price with
// @Description provided filters and period values (start, end),
// @Description if start or end undefined, returns sum for all the time.
// @Tags subs
// @Router /subs/sum [get]
// @Param name query string false "Sub's service name" Example(Spotify)
// @Param start query string false "Start period" Example(01-2015)
// @Param end query string false "End period" Example(03-2016)
// @Param uid query string false "User ID" Example(60601fee-2bf1-4721-ae6f-7636e79a0cba)
// @Produce json
// @Success 200 {object} sumResponse
// @Failure 500 {object} map[string]string
func (s *Server) getPriceSum(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	reqID := r.Context().Value("Request-ID").(string)
	filter := getFilterFromQuery(r)
	period, err := getPeriodFromQuery(r)
	if err != nil {
		slog.Error("sum request with invalid period dates",
			slog.String("req_id", reqID),
			slog.String("from", r.RemoteAddr))
		writeErrorMessage(w, http.StatusBadRequest, errvalues.ErrInvalidRequest)
		return
	}
	sum, err := s.subsRepo.PriceSum(filter, period)
	if err != nil {
		slog.Error("getting subs sum error",
			slog.String("error", err.Error()),
			slog.String("req_id", reqID),
			slog.String("from", r.RemoteAddr))
		writeErrorMessage(w, http.StatusInternalServerError, errvalues.ErrInternal)
		return
	}
	err = sonic.ConfigFastest.NewEncoder(w).Encode(sumResponse{
		Sum: sum,
	})
	if err != nil {
		slog.Error("error providing result",
			slog.String("error", err.Error()),
			slog.String("req_id", reqID),
			slog.String("from", r.RemoteAddr))
		writeErrorMessage(w, http.StatusInternalServerError, errvalues.ErrInternal)
		return
	}
	slog.Info("successfully provided subscriptions' price sum",
		slog.String("req_id", reqID),
		slog.String("from", r.RemoteAddr))
}
