package domain

type (
	Chunk struct {
		UploadID      string // unique id for the current upload.
		ChunkNumber   int64
		TotalChunks   int64
		TotalFileSize int64 // in bytes
		Filename      string
		Data          []byte
	}
)
