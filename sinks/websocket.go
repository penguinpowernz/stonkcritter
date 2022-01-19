package sinks

import (
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/penguinpowernz/stonkcritter/models"
)

// Websockets will create a sink that will send full disclosure objects as JSON
// to connected websocket clients.  The wsURL should contain the address to bind
// to for the websockets server, and the path (e.g. 0.0.0.0:8080/ws/trades).  A
// new HTTP server will be created solely for this sink.
func Websockets(wsURL string) (Sink, error) {
	var conns []*websocket.Conn
	connlock := new(sync.RWMutex)

	path := "/"
	bits := strings.Split(wsURL, "/")
	bind := bits[0]
	if len(bits) == 2 {
		path = "/" + strings.Join(bits[1:], "/")
	}

	r := gin.Default()
	r.GET(path, func(c *gin.Context) {
		conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("Failed to set websocket upgrade for %s: %+v", c.Request.RemoteAddr, err)
			return
		}

		connlock.Lock()
		defer connlock.Unlock()
		conns = append(conns, conn)
	})

	var err error
	go func() {
		err = r.Run(bind)
	}()
	time.Sleep(time.Second / 4)
	if err != nil {
		return nil, err
	}

	return func(d models.Disclosure) error {
		connlock.RLock()
		defer connlock.RUnlock()

		for i, c := range conns {
			if c == nil {
				continue
			}

			if err := c.WriteJSON(d); err == websocket.ErrCloseSent || err == io.EOF {
				conns[i] = nil
			}
		}

		return nil
	}, nil
}

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
