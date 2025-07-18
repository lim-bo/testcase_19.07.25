package api

import (
	"log/slog"
	"net/http"
	"testcase/models"

	"github.com/go-chi/chi/v5"
)

type SubsRepository interface {
	AddSub(s *models.Subscription) error
	GetSub(id int) (*models.Subscription, error)
	UpdateSub(id int, s *models.Subscription) error
	DeleteSub(id int) error
	ListSubs(opts *models.ListOpts) ([]*models.Subscription, error)
	PriceSum(filter map[string]interface{}) (int, error)
}

type Server struct {
	mx       *chi.Mux
	subsRepo SubsRepository
}

func New(sr SubsRepository) *Server {
	return &Server{
		mx:       chi.NewMux(),
		subsRepo: sr,
	}
}

func (s *Server) mountEndpoints() {
	s.mx.Use(s.CORSMiddleware, s.RequestIDMiddleware)
	s.mx.Route("/subs", func(r chi.Router) {
		r.Post("/add", s.addSubscription)
		r.Route("/{id}", func(r chi.Router) {
			r.Use(s.subIDMiddleware)
			r.Get("/", s.getSubscription)
			r.Put("/", s.updateSubscription)
			r.Delete("/", s.deleteSubscription)
		})
		r.Get("/list", s.listSubscriptions)
		r.Get("/sum", s.getPriceSum)
	})
}

func (s *Server) Run(address string) error {
	s.mountEndpoints()
	slog.Info("server is running on " + address)
	return http.ListenAndServe(address, s.mx)
}
