package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"

	"cloud.google.com/go/storage"
)

// uploadAudioFile uploads the audio file to Google Cloud Storage.
func uploadAudioFile(ctx context.Context, bucketName, objectName, filePath string) error {
	// Open the local file.
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("os.Open: %w", err)
	}
	defer f.Close()

	// Create a new Cloud Storage client.
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %w", err)
	}
	defer client.Close()

	// Create a new bucket handle.
	bucket := client.Bucket(bucketName)

	// Create a new object handle.
	object := bucket.Object(objectName)

	// Create a new writer.
	w := object.NewWriter(ctx)

	// Set the Content-Type header.
	w.ContentType = mime.TypeByExtension(filepath.Ext(filePath))

	// Copy the file to the writer.
	if _, err := io.Copy(w, f); err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}

	// Close the writer.
	if err := w.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %w", err)
	}

	return nil
}

// objectExists checks if an object exists in Google Cloud Storage.
func objectExists(ctx context.Context, bucketName, objectName string) (bool, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return false, fmt.Errorf("storage.NewClient: %w", err)
	}
	defer client.Close()

	_, err = client.Bucket(bucketName).Object(objectName).Attrs(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("Object(%q).Attrs: %w", objectName, err)
	}

	return true, nil
}

// deleteObject deletes an object from Google Cloud Storage.
func deleteObject(ctx context.Context, bucketName, objectName string) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %w", err)
	}
	defer client.Close()

	o := client.Bucket(bucketName).Object(objectName)
	if err := o.Delete(ctx); err != nil {
		return fmt.Errorf("Object(%q).Delete: %w", objectName, err)
	}

	return nil
}
