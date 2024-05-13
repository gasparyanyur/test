package repository

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"

	"node-test/internal/domain"
)

type (
	nodeRepository struct {
		fs *gridfs.Bucket
	}

	NodeRepository interface {
		State(ctx context.Context) (int64, error)
		Add(file *domain.Chunk) error
	}
)

// NewNodeRepository creates a new NodeRepository instance.
func NewNodeRepository(database *mongo.Database) (NodeRepository, error) {
	fs, err := gridfs.NewBucket(
		database,
		options.GridFSBucket().SetName("fs"),
	)
	if err != nil {
		return nil, err
	}
	return &nodeRepository{fs: fs}, nil
}

// State returns the total size of all uploaded files in bytes.
func (repo *nodeRepository) State(ctx context.Context) (int64, error) {

	state, err := repo.fs.GetChunksCollection().CountDocuments(ctx, bson.D{})
	if err != nil {
		return 0, err
	}

	state = state * 255 * 1024

	return state, nil
}

// RetrieveFileByUploadID retrieves a file chunk from GridFS by its upload ID.
func (repo *nodeRepository) RetrieveFileByUploadID(ctx context.Context, uploadID string) (*domain.Chunk, error) {

	cursor, err := repo.fs.FindContext(ctx, bson.D{{Key: "metadata.UploadID", Value: uploadID}})
	if err != nil {
		return nil, fmt.Errorf("failed to find file in GridFS: %v", err)
	}
	defer cursor.Close(ctx)

	// Iterate over the cursor to retrieve the file.
	if cursor.Next(ctx) {
		var chunk domain.Chunk
		err := cursor.Decode(&chunk)
		if err != nil {
			return nil, fmt.Errorf("failed to decode file chunk: %v", err)
		}
		return &chunk, nil
	}

	return nil, fmt.Errorf("file with upload ID %s not found", uploadID)
}

func (repo *nodeRepository) Add(file *domain.Chunk) error {

	fsFileName := fmt.Sprintf("%s_%v", file.Filename, file.ChunkNumber)
	opts := &options.UploadOptions{}
	opts.SetMetadata(map[string]interface{}{
		"UploadID":      file.UploadID,
		"ChunkNumber":   file.ChunkNumber,
		"TotalChunks":   file.TotalChunks,
		"TotalFileSize": file.TotalFileSize,
		"Filename":      file.Filename,
	})
	uploadStream, err := repo.fs.OpenUploadStream(fsFileName, opts)
	if err != nil {
		return fmt.Errorf("failed to open upload stream: %w", err)
	}
	defer uploadStream.Close()
	// Write chunk data to GridFS.
	_, err = uploadStream.Write(file.Data)
	if err != nil {
		return fmt.Errorf("failed to write data to upload stream: %w", err)
	}

	return nil

}
