package http

import (
	"context"
	"encoding/json"
	"github.com/getsentry/sentry-go"
	"github.com/superhorsy/quest-app-backend/internal/core/helpers"
	"net/http"
	"strings"

	"github.com/superhorsy/quest-app-backend/internal/core/errors"
	"github.com/superhorsy/quest-app-backend/internal/core/logging"
	"go.uber.org/zap"
)

type ErrorMessage struct {
	Error string `json:"error"`
}

func handleError(ctx context.Context, w http.ResponseWriter, err error) {
	logging.From(ctx).Error("error occurred in request", zap.Error(err))

	switch {
	case errors.Is(err, errors.ErrInvalidRequest):
		fallthrough
	case errors.Is(err, errors.ErrValidation):
		w.WriteHeader(http.StatusBadRequest)
	case errors.Is(err, errors.ErrNotFound):
		w.WriteHeader(http.StatusNotFound)
	case errors.Is(err, errors.ErrForbidden):
		w.WriteHeader(http.StatusForbidden)
	case errors.Is(err, errors.ErrUnknown):
		fallthrough
	default:
		w.WriteHeader(http.StatusInternalServerError)
		s := helpers.GetConfig(ctx).SentryDSN
		if s != "" {
			sentry.CaptureException(err)
		}
	}

	// TODO we may need to strip additional error information
	errorMessage := strings.Split(err.Error(), errors.ErrSeperator)[0]

	data, err := json.Marshal(ErrorMessage{
		Error: errorMessage,
	})
	if err != nil {
		logging.From(ctx).Error("failed to serialize error response", zap.Error(err))
		data = []byte(`{"error": "internal server error"}`)
	}

	_, err = w.Write(data)
	if err != nil {
		logging.From(ctx).Error("failed to write error response", zap.Error(err))
	}
}
