package http

type (
	Chunk struct {
		UploadID      string `json:"upload_id" validate:"required"`
		ChunkNumber   int64  `json:"chunk_number" validate:"required"`
		TotalChunks   int64  `json:"total_chunks" validate:"required"`
		TotalFileSize int64  `json:"total_file_size" validate:"required"`
		Filename      string `json:"filename" validate:"required"`
		Data          []byte `json:"data" validate:"required"`
	}

	ChunkMetadata struct {
		TotalFileSize int64  `json:"total_file_size" validate:"required"`
		Filename      string `json:"filename" validate:"required"`
	}
)
