package file_storage

import (
	"context"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

type LocalFileStorage struct {
}

// New will instantiate a new instance of RecordStore.
func New() *LocalFileStorage {
	return &LocalFileStorage{}
}

func (s *LocalFileStorage) Upload(_ context.Context, file multipart.File, filename string) error {
	path := filepath.Join(".", "files")
	_ = os.MkdirAll(path, os.ModePerm)

	fullPath := path + "/" + filename

	dest, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer dest.Close()

	// Copy the file to the destination path
	_, err = io.Copy(dest, file)
	if err != nil {
		return err
	}
	return nil
}
