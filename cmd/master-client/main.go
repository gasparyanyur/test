package main

import (
	"errors"
	"flag"
	"io"
	"log"
	"net/url"
	"os"

	"github.com/gorilla/websocket"

	"node-test/internal/common/http"
)

const (
	chunkSize = 50 * 1024 //50KB
)

func main() {

	actionName := flag.String("action", "", "action name [upload,download]")
	objName := flag.String("obj", "", "file name or id")
	host := flag.String("host", "127.0.0.1:8080", "master server")

	flag.Parse()

	if actionName == nil {
		panic(errors.New("action hasn't been provided"))
	}

	if objName == nil {
		panic(errors.New("object hasn't been provided"))
	}

	switch *actionName {
	case "upload":
		err := upload(*objName, *host)
		if err != nil {
			log.Fatal(err)
		}
	case "download":
		download(*objName, *host)
	default:
		panic(errors.New("unregistered action name"))

	}

}

func upload(fileName, host string) error {

	log.Println("uploading file")

	u := url.URL{Scheme: "ws", Host: host, Path: "/api/v1/storage/ws/upload"}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer conn.Close()

	f, err := os.Open(fileName)
	if err != nil {
		return err
	}

	stat, err := f.Stat()
	if err != nil {
		return err
	}

	metadata := &http.ChunkMetadata{
		TotalFileSize: stat.Size(),
		Filename:      fileName,
	}
	err = conn.WriteJSON(metadata)
	if err != nil {
		return err
	}

	buf := make([]byte, 0, chunkSize)
	for {
		n, err := f.Read(buf[:cap(buf)])
		buf = buf[:n]
		if n == 0 {
			if err == nil {
				log.Println(err)
				continue
			}
			if err == io.EOF {
				break
			}
		}
		err = conn.WriteMessage(websocket.BinaryMessage, buf)
		if err != nil {
			log.Println(err)
			break
		}
	}

	return nil

}

func download(fileName, host string) {}
