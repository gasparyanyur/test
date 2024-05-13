package rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"

	commonHttp "node-test/internal/common/http"
	"node-test/internal/domain"
	"node-test/internal/master/service"
)

const (
	maxChunkSize = 50 * 1024
)

type (
	storageHandler struct {
		service service.UploadService
		socket  websocket.Upgrader
	}
)

func newStorageHandler(storageService service.UploadService) *storageHandler {
	return &storageHandler{
		service: storageService,
		socket:  websocket.Upgrader{},
	}
}

func (h *storageHandler) WSUpload(c echo.Context) error {

	ws, err := h.socket.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	ctx := c.Request().Context()

	uploadChan := h.service.UploadChunkedAsync(ctx)
	defer close(uploadChan)

	var msg []byte
	_, msg, err = ws.ReadMessage()
	if err != nil {
		return err
	}
	var metadata commonHttp.ChunkMetadata
	err = json.Unmarshal(msg, &metadata)
	if err != nil {
		return err
	}

	var (
		totalChunkNum       = metadata.TotalFileSize / maxChunkSize
		chunkNum      int64 = 1
		chunkSent     int64 = 0
		uploadID            = uuid.New().String()
	)

	if metadata.TotalFileSize%maxChunkSize > 0 {
		totalChunkNum++
	}

upload:
	for {
		select {
		case <-ctx.Done():
			break
		default:
			_, chunk, err := ws.ReadMessage()
			if err != nil {
				break upload
			}

			chunkSent += int64(len(chunk))
			chunkNum++

			if chunkSent > metadata.TotalFileSize {
				err = fmt.Errorf(
					"wrong count of data received! expected %v actual %v",
					metadata.TotalFileSize,
					chunkSent)

				break upload
			}

			uploadChan <- &domain.Chunk{
				UploadID:      uploadID,
				ChunkNumber:   chunkNum,
				TotalChunks:   totalChunkNum,
				TotalFileSize: metadata.TotalFileSize,
				Filename:      metadata.Filename,
				Data:          chunk,
			}
		}
	}

	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func (h *storageHandler) WSDownload(c echo.Context) error {

	ws, err := h.socket.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	ctx := c.Request().Context()

	downloadList, err := h.service.DownloadChunked(ctx)

	return c.NoContent(http.StatusOK)
}
