//go:generate mockgen -destination=./mocks/http_mock.go -package mocks github.com/superhorsy/quest-app-backend/internal/transport/http Users,DB

package http

import (
	"context"
	"encoding/json"
	questModel "github.com/superhorsy/quest-app-backend/internal/quests/model"
	"net/http"

	"github.com/goji/httpauth"
	"github.com/gorilla/mux"
	"github.com/superhorsy/quest-app-backend/internal/users/model"
)

// Users represents a type that can provide CRUD operations on users.
type Users interface {
	CreateUser(ctx context.Context, user *model.UserWithPass) (*model.User, error)
	GetUser(ctx context.Context, id string) (*model.User, error)
	FindUsers(ctx context.Context, filters []model.Filter, offset, limit int64) ([]*model.User, error)
	UpdateUser(ctx context.Context, user *model.UserWithPass) (*model.User, error)
	DeleteUser(ctx context.Context, id string) error
}

// Quests represents a type that can provide CRUD operations on quests.
type Quests interface {
	CreateQuest(ctx context.Context, quest *questModel.Quest) (*questModel.Quest, error)
	GetQuest(ctx context.Context, id string) (*questModel.Quest, error)
	UpdateQuest(ctx context.Context, quest *questModel.Quest) (*questModel.Quest, error)
	GetQuestsByUser(ctx context.Context, uuid string, offset int, limit int) ([]questModel.Quest, error)
}

// DB represents a type that can be used to interact with the database.
type DB interface {
	PingContext(ctx context.Context) error
}

// Server represents an HTTP server that can handle requests for this microservice.
type Server struct {
	users  Users
	quests Quests
	db     DB
}

// New will instantiate a new instance of Server.
func New(u Users, q Quests, db DB) *Server {
	return &Server{
		users:  u,
		quests: q,
		db:     db,
	}
}

// AddRoutes will add the routes this server supports to the router.
func (s *Server) AddRoutes(r *mux.Router) error {
	r.HandleFunc("/health", s.healthCheck).Methods(http.MethodGet)

	r = r.PathPrefix("/v1").Subrouter()

	authHandler := httpauth.SimpleBasicAuth("test", "test")

	r.HandleFunc("/login", s.login).Methods(http.MethodPost)

	r.Use(authHandler)
	r.Use(EnforceJSONHandler)
	r.Use(JsonResponse)
	r.Use(AllowCors)

	r.HandleFunc("/quests", s.createQuest).Methods(http.MethodPost)
	r.HandleFunc("/quests/{id}", s.getQuest).Methods(http.MethodGet)
	r.HandleFunc("/quests/{id}", s.updateQuest).Methods(http.MethodPut)
	r.HandleFunc("/quests/created", s.getQuestsByUser).Methods(http.MethodGet)

	r.HandleFunc("/user", s.createUser).Methods(http.MethodPost)
	r.HandleFunc("/user/{id}", s.getUser).Methods(http.MethodGet)
	r.HandleFunc("/user/{id}", s.updateUser).Methods(http.MethodPut)
	r.HandleFunc("/user/{id}", s.deleteUser).Methods(http.MethodDelete)

	// Not the most RESTful way of doing this as it won't really be cachable but provides easier parsing of the inputs for now
	r.HandleFunc("/users/search", s.searchUsers).Methods(http.MethodPost)

	return nil
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	if err := s.db.PingContext(r.Context()); err != nil {
		handleError(r.Context(), w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleResponse(ctx context.Context, w http.ResponseWriter, data interface{}) {
	jsonRes := struct {
		Data interface{} `json:"data"`
	}{
		Data: data,
	}

	dataBytes, err := json.Marshal(jsonRes)
	if err != nil {
		handleError(ctx, w, err)
		return
	}

	if _, err := w.Write(dataBytes); err != nil {
		handleError(ctx, w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
