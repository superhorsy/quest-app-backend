package media

import (
	"context"
	"fmt"
	"github.com/superhorsy/quest-app-backend/internal/events"
	fileStorage "github.com/superhorsy/quest-app-backend/internal/media/file_storage"
	"github.com/superhorsy/quest-app-backend/internal/media/model"
	"github.com/superhorsy/quest-app-backend/internal/media/store"
	"mime/multipart"
	"os"
	"path/filepath"
)

// Events represents a type for producing events on user CRUD operations.
type Events interface {
	Produce(ctx context.Context, topic events.Topic, payload interface{})
}

type Media struct {
	recordStore store.RecordStore
	fileStorage fileStorage.LocalFileStorage
	events      Events
}

func New(rs *store.RecordStore, fs *fileStorage.LocalFileStorage, e Events) *Media {
	return &Media{
		recordStore: *rs,
		fileStorage: *fs,
		events:      e,
	}
}

func (m Media) UploadFile(ctx context.Context, file multipart.File, filename string, mediaType model.MediaType) (*model.MediaRecord, error) {
	record := &model.MediaRecord{
		Storage: "local",
		Type:    mediaType,
	}

	record, err := m.recordStore.InsertMedia(ctx, record)
	if err != nil {
		return nil, err
	}

	ext := filepath.Ext(filename)
	filename = fmt.Sprintf("%s%s", record.ID, ext)

	err = m.fileStorage.Upload(ctx, file, filename)

	record.Filename = filename

	staticFilesEndpoint := os.Getenv("STATIC_FILES_ENDPOINT")
	if staticFilesEndpoint == "" {
		staticFilesEndpoint = "/files/"
	}
	record.Link = fmt.Sprintf("%s%s", staticFilesEndpoint, filename)

	record, err = m.recordStore.UpdateMedia(ctx, record)
	if err != nil {
		return nil, err
	}

	m.events.Produce(ctx, events.TopicMedia, events.MediaEvent{
		EventType:   events.EventTypeUserCreated,
		ID:          record.ID,
		MediaRecord: record,
	})

	return record, nil
}

func (m Media) GetMedia(ctx context.Context, id string) (*model.MediaRecord, error) {
	media, err := m.recordStore.GetMedia(ctx, id)
	if err != nil {
		return nil, err
	}
	return media, nil
}
