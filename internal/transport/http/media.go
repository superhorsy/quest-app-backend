package http

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/superhorsy/quest-app-backend/internal/core/errors"
	"github.com/superhorsy/quest-app-backend/internal/core/helpers"
	"github.com/superhorsy/quest-app-backend/internal/core/logging"
	"github.com/superhorsy/quest-app-backend/internal/media/model"
	"go.uber.org/zap"
	"net/http"
	"path/filepath"
)

type MediaRequest struct {
	Type model.MediaType `json:"type"`
}

func (s *Server) uploadMedia(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	fmt.Println("File Upload Endpoint Hit")

	// Parse our multipart form, 5 << 20 specifies a maximum
	// upload of 5 MB files.
	// left shift 5 << 20 which results in 5*2^20
	// x << y, results in x*2^y
	err := r.ParseMultipartForm(5 << 20)
	if err != nil {
		logging.From(ctx).Error("failed to upload file: file exceeds 5Mb", zap.Error(err))
		handleError(ctx, w, errors.ErrInvalidRequest.Wrap(err))
		return
	}

	t := model.MediaType(r.Form.Get("type"))
	if !(t == model.Sound || t == model.Image) {
		err = errors.New("bad file type")
		logging.From(ctx).Error(err.Error(), zap.Error(err))
		handleError(ctx, w, errors.ErrInvalidRequest.Wrap(err))
		return
	}

	mr := MediaRequest{
		Type: t,
	}

	// FormFile returns the first file for the given key `myFile`
	// it also returns the FileHeader, so we can get the Filename,
	// the Header and the size of the file
	file, header, err := r.FormFile("file")
	if err != nil {
		logging.From(ctx).Error("Error Retrieving the File", zap.Error(err))
		handleError(ctx, w, errors.ErrInvalidRequest.Wrap(err))
		return
	}
	defer file.Close()

	imagesExt := []string{
		".png",
		".jpg",
		".jpeg",
	}
	soundExt := []string{
		".mp3",
		".wav,",
		".webm",
	}

	fmt.Printf("Uploaded File: %+v", header.Filename)
	fmt.Printf("File Size: %+v\n", header.Size)
	fmt.Printf("MIME Header: %+v\n", header.Header)

	ext := filepath.Ext(header.Filename)

	if !(helpers.SliceContains(imagesExt, ext) || helpers.SliceContains(soundExt, ext)) {
		err = errors.New(fmt.Sprintf("bad file type: %s", ext))
		logging.From(ctx).Error(err.Error(), zap.Error(err))
		handleError(ctx, w, errors.ErrInvalidRequest.Wrap(err))
		return
	}

	if helpers.SliceContains(imagesExt, ext) && t != model.Image {
		err = errors.New(fmt.Sprintf("bad file type for type 'image': %s", ext))
		logging.From(ctx).Error(err.Error(), zap.Error(err))
		handleError(ctx, w, errors.ErrInvalidRequest.Wrap(err))
		return
	}

	if helpers.SliceContains(soundExt, ext) && t != model.Sound {
		err = errors.New(fmt.Sprintf("bad file type for type 'sound': %s", ext))
		logging.From(ctx).Error(err.Error(), zap.Error(err))
		handleError(ctx, w, errors.ErrInvalidRequest.Wrap(err))
		return
	}

	mediaRecord, err := s.media.UploadFile(ctx, file, header.Filename, mr.Type)
	if err != nil {
		logging.From(ctx).Error("failed to upload media record", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	handleResponse(ctx, w, mediaRecord)
}

func (s *Server) getMedia(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		err := errors.New(fmt.Sprintf("bad id: %s", id))
		logging.From(ctx).Error(err.Error(), zap.Error(err))
		handleError(ctx, w, errors.ErrInvalidRequest.Wrap(err))
		return
	}

	media, err := s.media.GetMedia(ctx, id)
	if err != nil {
		logging.From(ctx).Error("failed to fetch media", zap.Error(err))
		handleError(ctx, w, err)
		return
	}

	handleResponse(ctx, w, media)
}
