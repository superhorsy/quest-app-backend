//go:generate mockgen -destination=./mocks/http_mock.go -package mocks github.com/superhorsy/quest-app-backend/internal/transport/http Users,DB

package http

import (
	"context"
	"encoding/json"
	"github.com/superhorsy/quest-app-backend/internal/core/errors"
	"github.com/superhorsy/quest-app-backend/internal/core/logging"
	mediaModel "github.com/superhorsy/quest-app-backend/internal/media/model"
	questModel "github.com/superhorsy/quest-app-backend/internal/quests/model"
	"go.uber.org/zap"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/superhorsy/quest-app-backend/internal/users/model"
)

// Users represents a type that can provide CRUD operations on users.
type Users interface {
	CreateUser(ctx context.Context, user *model.UserWithPass) (*model.User, error)
	GetUser(ctx context.Context, id string) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	FindUsers(ctx context.Context, filters []model.Filter, offset, limit int64) ([]*model.User, error)
	UpdateUser(ctx context.Context, user *model.UserWithPass) (*model.User, error)
	DeleteUser(ctx context.Context, id string) error
}

// Quests represents a type that can provide CRUD operations on quests.
type Quests interface {
	CreateQuest(ctx context.Context, quest *questModel.QuestWithSteps) (*questModel.QuestWithSteps, error)
	GetQuest(ctx context.Context, id string) (*questModel.QuestWithSteps, error)
	UpdateQuest(ctx context.Context, quest *questModel.QuestWithSteps) (*questModel.QuestWithSteps, error)
	GetQuestsByUser(ctx context.Context, uuid string, offset int, limit int) ([]questModel.Quest, *questModel.Meta, error)
	GetQuestsAvailable(ctx context.Context, email string, offset int, limit int) ([]questModel.QuestAvailable, *questModel.Meta, error)
	DeleteQuest(ctx context.Context, id string) error
	CreateAssignment(ctx context.Context, request questModel.SendQuestRequest) error
	GetAssignment(ctx context.Context, questId string, email *string) (*questModel.Assignment, error)
	UpdateAssignment(ctx context.Context, questId string, email *string, currentStep int, status questModel.Status) error
}

type Media interface {
	UploadFile(ctx context.Context, file multipart.File, filename string, mediaType mediaModel.MediaType) (*mediaModel.MediaRecord, error)
	GetMedia(ctx context.Context, id string) (*mediaModel.MediaRecord, error)
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
	media  Media
}

// New will instantiate a new instance of Server.
func New(u Users, q Quests, db DB, m Media) *Server {
	return &Server{
		users:  u,
		quests: q,
		media:  m,
		db:     db,
	}
}

// AddRoutes will add the routes this server supports to the router.
func (s *Server) AddRoutes(r *mux.Router) error {

	// Health check
	r.HandleFunc("/health", s.healthCheck).Methods(http.MethodGet)

	r = r.PathPrefix("/api").Subrouter()
	r = r.PathPrefix("/v1").Subrouter()

	//Authorisation
	auth := r.Name("auth").Subrouter()
	auth.Use(JsonResponse)
	auth.Use(EnforceJSONHandler)
	auth.Path("/login").Handler(http.HandlerFunc(s.login)).Methods(http.MethodPost)
	auth.Path("/register").Handler(http.HandlerFunc(s.register)).Methods(http.MethodPost)

	// Media handler
	media := r.Name("media").Subrouter()
	media.Use(authHandler)
	media.Use(JsonResponse)
	media.HandleFunc("/media/upload", s.uploadMedia).Methods(http.MethodPost)
	media.HandleFunc("/media/{id}", s.getMedia).Methods(http.MethodGet)

	api := r.Name("api").Subrouter()
	api.Use(authHandler)
	api.Use(JsonResponse)
	api.Use(EnforceJSONHandler)

	// Profile
	api.HandleFunc("/profile", s.getCurrentUser).Methods(http.MethodGet)
	// Quests
	api.HandleFunc("/quests", s.createQuest).Methods(http.MethodPost)
	api.HandleFunc("/quests/created", s.getQuestsByUser).Methods(http.MethodGet)
	api.HandleFunc("/quests/available", s.getAvailableQuests).Methods(http.MethodGet)
	api.HandleFunc("/quests/{id}", s.getQuest).Methods(http.MethodGet)
	api.HandleFunc("/quests/{id}", s.updateQuest).Methods(http.MethodPut)
	api.HandleFunc("/quests/{id}", s.deleteQuest).Methods(http.MethodDelete)
	api.HandleFunc("/quests/{id}/send", s.sendQuest).Methods(http.MethodPost)
	api.HandleFunc("/quests/{id}/start", s.startQuest).Methods(http.MethodPost)
	api.HandleFunc("/quests/{id}/next", s.checkAnswer).Methods(http.MethodPost)
	api.HandleFunc("/quests/{id}/status", s.status).Methods(http.MethodGet)

	return nil
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	if err := s.db.PingContext(r.Context()); err != nil {
		handleError(r.Context(), w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type Response struct {
	Data interface{} `json:"data"`
	Meta interface{} `json:"meta,omitempty"`
	Jwt  string      `json:"jwt,omitempty"`
}

func handleResponseWithMeta(ctx context.Context, w http.ResponseWriter, data interface{}, meta interface{}) {

	jsonRes := Response{
		Data: data,
		Meta: meta,
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

func handleResponse(ctx context.Context, w http.ResponseWriter, data interface{}, jwt ...string) {

	jsonRes := Response{
		Data: data,
	}

	if len(jwt) > 0 {
		jsonRes.Jwt = jwt[0]
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

func parseBodyIntoStruct[K any](r *http.Request, target K) (*K, error) {
	ctx := r.Context()
	data, err := io.ReadAll(r.Body)
	if err != nil {
		logging.From(ctx).Error("failed to read request body", zap.Error(err))
		return nil, errors.ErrUnknown.Wrap(err)
	}

	if err := json.Unmarshal(data, &target); err != nil {
		logging.From(ctx).Error("failed to unmarshal json body", zap.Error(err))
		return nil, errors.ErrInvalidRequest.Wrap(err)
	}

	return &target, nil
}
