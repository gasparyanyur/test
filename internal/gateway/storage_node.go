package gateway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"

	commonRest "node-test/internal/common/http"
	"node-test/internal/common/pool"
	"node-test/internal/domain"
	"node-test/internal/master/config"
)

const (
	minStorageNodeCount = 6

	nodeStatePath            = "/state"
	nodeStateValueHeaderName = "X-NODE-STATE"
)

type (
	nodes []*nodeState

	nodeState struct {
		ip           string
		currentState int64
		sync.RWMutex
	}

	storageNodeGateway struct {
		nodes nodes
		cfg   config.StorageConfig
		pool  *pool.Pool
		sync.RWMutex
	}

	StorageNodeGateway interface {
		SendAsync(data *domain.Chunk)
		DownloadAsync(id string) chan *domain.Chunk
	}

	sendAsyncJob struct {
		url  string
		data *commonRest.Chunk
	}

	downloadAsyncJob struct {
		url   string
		id    string
		chann chan *domain.Chunk
	}
)

func NewStorageNodeGateway(cfg config.StorageConfig, pool *pool.Pool) (StorageNodeGateway, error) {

	fsNodes := make([]*nodeState, 0, len(cfg.Nodes))

	gateway := &storageNodeGateway{
		pool: pool,
	}

	for _, ip := range cfg.Nodes {
		state, err := gateway.loadState(ip)
		if err != nil {
			return nil, fmt.Errorf("can't determine the current state of node %v %w", ip, err)
		}
		fsNodes = append(fsNodes, &nodeState{
			ip:           ip,
			currentState: state,
		})
	}

	//if len(fsNodes) < minStorageNodeCount {
	//	return nil, errors.New("not enough storage nodes")
	//}

	gateway.nodes = fsNodes
	gateway.cfg = cfg

	return gateway, nil
}

func (g *storageNodeGateway) loadState(ip string) (int64, error) {
	g.Lock()
	defer g.Unlock()

	stateResponse, err := http.Get(ip + nodeStatePath)
	if err != nil {
		return 0, err
	}

	stateStr := stateResponse.Header.Get(nodeStateValueHeaderName)

	state, err := strconv.Atoi(stateStr)
	if err != nil {
		return 0, err
	}

	return int64(state), nil
}

// balanceStates sorting nodes by their current state
func (g *storageNodeGateway) balanceStates() {
	g.Lock()
	defer g.Unlock()
	sort.Slice(g.nodes, func(i, j int) bool {
		g.nodes[i].Lock()
		defer g.nodes[i].Unlock()
		ext := g.nodes[i].currentState < g.nodes[j].currentState
		return ext
	})

	return
}

// increase increases current state of the specific node
func (s *nodeState) increase(len int64) {
	s.Lock()
	defer s.Unlock()
	s.currentState = s.currentState + len
	return
}

// getUnloaded retrieves the node with more available space
func (g *storageNodeGateway) getUnloaded() *nodeState {
	g.RLock()
	defer g.RUnlock()
	first := g.nodes[0]
	return first
}

// SendAsync sends the async request to the node to store specific chunk
func (g *storageNodeGateway) SendAsync(data *domain.Chunk) {

	//  balance the state
	g.balanceStates()
	// get unloaded node
	node := g.getUnloaded()

	defer func() {
		node.increase(int64(len(data.Data)))
	}()

	g.pool.Submit(&sendAsyncJob{
		url: node.ip + "/upload",
		data: &commonRest.Chunk{
			UploadID:      data.UploadID,
			ChunkNumber:   data.ChunkNumber,
			TotalChunks:   data.TotalChunks,
			TotalFileSize: data.TotalFileSize,
			Filename:      data.Filename,
			Data:          data.Data,
		},
	})

}

func (g *storageNodeGateway) DownloadAsync(id string) chan *domain.Chunk {

	var downloadChunk = make(chan *domain.Chunk, 1)
	for _, node := range g.cfg.Nodes {
		g.pool.Submit(&downloadAsyncJob{
			url:   node,
			id:    id,
			chann: downloadChunk,
		})
	}

	return downloadChunk
}

func (j *sendAsyncJob) Do() error {

	body, err := json.Marshal(j.data)
	if err != nil {
		return fmt.Errorf("failed marshal data %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, j.url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create http request %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send http request %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad request status code %v", resp.StatusCode)
	}

	return nil
}

func (j *downloadAsyncJob) Do() error {

	u := url.URL{Scheme: "ws", Host: j.url, Path: "/api/v1/download"}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	err = conn.WriteMessage(websocket.TextMessage, []byte(j.id))
	if err != nil {
		return fmt.Errorf("failed to send request  %w", err)
	}

	for {
		var resp commonRest.Chunk
		if err := conn.ReadJSON(&resp); err != nil {
			log.Println(err)
			break
		}
		j.chann <- &domain.Chunk{
			UploadID:      j.id,
			ChunkNumber:   resp.ChunkNumber,
			TotalChunks:   resp.TotalChunks,
			TotalFileSize: resp.TotalFileSize,
			Filename:      resp.Filename,
			Data:          resp.Data,
		}
	}

	return nil
}
