package main

import (
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn *websocket.Conn
	done chan struct{}

	closeOnce sync.Once

	server *Server
}

//创建一个用户的API
func NewUser(conn *websocket.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string, 64),
		conn: conn,
		done: make(chan struct{}),

		server: server,
	}

	//启动监听当前user channel消息的goroutine
	go user.ListenMessage()

	return user
}

//用户的上线业务
func (this *User) Online() {

	//用户上线,将用户加入到onlineMap中
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	this.server.SetOnline(this)

	//广播当前用户上线消息
	this.server.BroadCastSystem(this, "已上线")
}

//用户的下线业务
func (this *User) Offline() {

	//用户下线,将用户从onlineMap中删除
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	this.server.SetOffline(this)

	//广播当前用户上线消息
	this.server.BroadCastSystem(this, "下线")

}

//给当前User对应的客户端发送消息
func (this *User) SendMsg(msg string) {
	select {
	case this.C <- msg:
	default:
		//避免慢客户端阻塞业务流程
	}
}

//用户处理消息的业务
func (this *User) DoMessage(msg string) {
	if msg == "PING" {
		this.SendMsg("PONG")
		return
	}

	if msg == "who" {
		//查询当前在线用户都有哪些

		this.server.mapLock.RLock()
		users := make([]*User, 0, len(this.server.OnlineMap))
		for _, user := range this.server.OnlineMap {
			users = append(users, user)
		}
		this.server.mapLock.RUnlock()

		for _, user := range users {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线..."
			this.SendMsg(onlineMsg)
		}

	} else if len(msg) > 7 && msg[:7] == "rename|" {
		//消息格式: rename|张三
		newName := strings.Split(msg, "|")[1]

		//判断name是否存在
		this.server.mapLock.Lock()
		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.server.mapLock.Unlock()
			this.SendMsg("当前用户名被使用")
			return
		}
		oldName := this.Name
		delete(this.server.OnlineMap, this.Name)
		this.server.OnlineMap[newName] = this
		this.server.mapLock.Unlock()

		this.Name = newName
		this.server.RenameOnline(oldName, newName, this.Addr)
		this.SendMsg("您已经更新用户名:" + this.Name)

	} else if len(msg) > 4 && msg[:3] == "to|" {
		//消息格式:  to|张三|消息内容

		//1 获取对方的用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			this.SendMsg("消息格式不正确，请使用 \"to|张三|你好啊\"格式。")
			return
		}

		//2 根据用户名 得到对方User对象
		this.server.mapLock.RLock()
		remoteUser, ok := this.server.OnlineMap[remoteName]
		this.server.mapLock.RUnlock()
		if !ok {
			this.SendMsg("该用户名不不存在")
			return
		}

		//3 获取消息内容，通过对方的User对象将消息内容发送过去
		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.SendMsg("无消息内容，请重发")
			return
		}
		remoteUser.SendMsg(this.Name + "对您说:" + content)
		this.server.EnqueueMessage(ChatMessage{
			FromUser:  this.Name,
			ToUser:    remoteName,
			Content:   content,
			CreatedAt: time.Now(),
		})

	} else {
		this.server.BroadCast(this, msg)
	}
}

//监听当前User channel的 方法,一旦有消息，就直接发送给对端客户端
func (this *User) ListenMessage() {
	for {
		select {
		case msg := <-this.C:
			this.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := this.conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
				return
			}
		case <-this.done:
			return
		}
	}
}

func (this *User) Close() {
	this.closeOnce.Do(func() {
		close(this.done)
		_ = this.conn.Close()
	})
}
