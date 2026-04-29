package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"IM-System/internal/config"
	"IM-System/internal/models"
	"IM-System/internal/pool"
	"IM-System/internal/redisstore"
)

type Server struct {
	Ip     string
	Port   int
	WSPath string

	//在线用户的列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	//消息广播的channel
	Message chan string

	Pool             *pool.ConnectionPool
	PersistQueue     chan models.ChatMessage
	Store            models.MessageStore
	PersistBatchSize int
	PersistInterval  time.Duration

	Redis            *redisstore.RedisStore
	HeartbeatTimeout time.Duration

	ctx context.Context
}

//创建一个server的接口
func NewServer(cfg config.Config, store models.MessageStore, redisStore *redisstore.RedisStore) *Server {
	server := &Server{
		Ip:                cfg.ServerIp,
		Port:              cfg.ServerPort,
		WSPath:            cfg.WSPath,
		OnlineMap:         make(map[string]*User),
		Message:           make(chan string, 128),
		Pool:              pool.NewConnectionPool(cfg.MaxConnections),
		PersistQueue:      make(chan models.ChatMessage, cfg.PersistQueueSize),
		Store:             store,
		PersistBatchSize:  cfg.PersistBatchSize,
		PersistInterval:   cfg.PersistInterval,
		Redis:             redisStore,
		HeartbeatTimeout:  cfg.HeartbeatTimeout,
		ctx:               context.Background(),
	}

	return server
}

//监听Message广播消息channel的goroutine，一旦有消息就发送给全部的在线User
func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message

		//将msg发送给全部的在线User
		this.mapLock.RLock()
		for _, cli := range this.OnlineMap {
			select {
			case cli.C <- msg:
			default:
				//避免慢客户端阻塞广播
			}
		}
		this.mapLock.RUnlock()
	}
}

//广播消息的方法
func (this *Server) BroadCast(user *User, msg string) {
	this.broadcast(user, msg, true)
}

func (this *Server) BroadCastSystem(user *User, msg string) {
	this.broadcast(user, msg, false)
}

func (this *Server) broadcast(user *User, msg string, persist bool) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	select {
	case this.Message <- sendMsg:
	default:
		//避免广播阻塞主链路
	}

	if !persist {
		return
	}

	this.EnqueueMessage(models.ChatMessage{
		FromUser:  user.Name,
		ToUser:    "",
		Content:   msg,
		CreatedAt: time.Now(),
	})
}

func (this *Server) EnqueueMessage(msg models.ChatMessage) {
	if this.PersistQueue == nil || this.Store == nil {
		return
	}
	select {
	case this.PersistQueue <- msg:
	default:
		fmt.Println("persist queue full, drop message")
	}
}

func (this *Server) SetOnline(user *User) {
	if this.Redis == nil {
		return
	}
	if err := this.Redis.SetOnline(this.ctx, user.Name, user.Addr); err != nil {
		fmt.Println("redis set online err:", err)
	}
}

func (this *Server) RefreshOnline(user *User) {
	if this.Redis == nil {
		return
	}
	if err := this.Redis.RefreshOnline(this.ctx, user.Name, user.Addr); err != nil {
		fmt.Println("redis refresh online err:", err)
	}
}

func (this *Server) SetOffline(user *User) {
	if this.Redis == nil {
		return
	}
	if err := this.Redis.SetOffline(this.ctx, user.Name); err != nil {
		fmt.Println("redis set offline err:", err)
	}
}

func (this *Server) RenameOnline(oldName string, newName string, addr string) {
	if this.Redis == nil {
		return
	}
	if err := this.Redis.RenameUser(this.ctx, oldName, newName, addr); err != nil {
		fmt.Println("redis rename err:", err)
	}
}
