package api

import (
	"context"
	"log/slog"
	"net/http"
	"testcase/models"

	_ "testcase/docs"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

type SubsRepository interface {
	AddSub(s *models.Subscription) error
	GetSub(id int) (*models.Subscription, error)
	UpdateSub(id int, s *models.Subscription) error
	DeleteSub(id int) error
	ListSubs(opts *models.ListOpts) ([]*models.Subscription, error)
	PriceSum(filter map[string]interface{}, period *models.RangeOpts) (int, error)
}

type Server struct {
	mx        *chi.Mux
	subsRepo  SubsRepository
	servEntry *http.Server
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
	s.mx.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))
}

func (s *Server) Run(address string) error {
	s.mountEndpoints()
	s.servEntry = &http.Server{
		Addr:    address,
		Handler: s.mx,
	}
	slog.Info("server is running on " + address)
	return s.servEntry.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.servEntry.Shutdown(ctx)
}
