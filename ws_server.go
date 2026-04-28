package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (this *Server) Start() error {
	go this.ListenMessager()
	StartPersistenceWorker(this.ctx, this.Store, this.PersistQueue, this.PersistBatchSize, this.PersistInterval)

	router := gin.Default()
	router.GET(this.WSPath, this.HandleWebSocket)

	addr := fmt.Sprintf("%s:%d", this.Ip, this.Port)
	return router.Run(addr)
}

func (this *Server) HandleWebSocket(c *gin.Context) {
	if this.Pool != nil && !this.Pool.Acquire() {
		c.String(http.StatusTooManyRequests, "server busy")
		return
	}

	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		if this.Pool != nil {
			this.Pool.Release()
		}
		return
	}
	defer func() {
		if this.Pool != nil {
			this.Pool.Release()
		}
	}()

	user := NewUser(conn, this)
	user.Online()
	defer func() {
		user.Offline()
		user.Close()
	}()

	isLive := make(chan struct{}, 1)
	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			msg := strings.TrimSpace(string(data))
			if msg == "" {
				continue
			}
			user.DoMessage(msg)
			this.RefreshOnline(user)
			select {
			case isLive <- struct{}{}:
			default:
			}
		}
	}()

	timer := time.NewTimer(this.HeartbeatTimeout)
	defer timer.Stop()

	for {
		select {
		case <-isLive:
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(this.HeartbeatTimeout)
		case <-timer.C:
			user.SendMsg("你被踢了")
			return
		case <-done:
			return
		}
	}
}
