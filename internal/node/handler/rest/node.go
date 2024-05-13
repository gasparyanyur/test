package rest

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"node-test/internal/common/errors"
	http2 "node-test/internal/common/http"
	"node-test/internal/domain"
	"node-test/internal/node/service"
)

type (
	nodeHandler struct {
		nodeService service.NodeService
	}
)

const (
	stateResponseHeaderName = "X-NODE-STATE"
)

func newNodeHandler(nodeService service.NodeService) *nodeHandler {
	return &nodeHandler{
		nodeService: nodeService,
	}
}

func (h *nodeHandler) State(c echo.Context) error {

	nodeState, err := h.nodeService.State(context.Background())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errors.NewInternalError(err))
	}

	c.Response().Header().Set(stateResponseHeaderName, fmt.Sprintf("%v", nodeState.Free))
	return c.NoContent(http.StatusOK)
}

func (h *nodeHandler) Upload(c echo.Context) error {

	var request http2.Chunk
	if err := c.Bind(&request); err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadRequest, errors.NewInternalError(err))
	}

	fmt.Println("received request", request.UploadID, request.ChunkNumber)

	if err := h.nodeService.Upload(&domain.Chunk{
		UploadID:      request.UploadID,
		ChunkNumber:   request.ChunkNumber,
		TotalChunks:   request.TotalChunks,
		TotalFileSize: request.TotalFileSize,
		Filename:      request.Filename,
		Data:          request.Data,
	}); err != nil {
		return c.JSON(http.StatusInternalServerError, errors.NewInternalError(err))
	}

	return c.NoContent(http.StatusOK)

}
