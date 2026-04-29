package client

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	ServerIp   string
	ServerPort int
	WSPath     string
	Name       string
	conn       *websocket.Conn
	flag       int //当前client的模式
	writeMu    sync.Mutex
}

func NewClient(serverIp string, serverPort int, wsPath string) *Client {
	//创建客户端对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		WSPath:     wsPath,
		flag:       999,
	}

	//链接server
	wsURL := fmt.Sprintf("ws://%s:%d%s", serverIp, serverPort, wsPath)
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		fmt.Println("websocket dial error:", err)
		return nil
	}

	client.conn = conn

	//返回对象
	return client
}

// 处理server回应的消息， 直接显示到标准输出即可
func (client *Client) DealResponse() {
	for {
		_, data, err := client.conn.ReadMessage()
		if err != nil {
			fmt.Fprintln(os.Stdout, "连接已断开")
			return
		}
		msg := string(data)
		if msg == "PONG" {
			continue
		}
		fmt.Fprintln(os.Stdout, msg)
	}
}

func (client *Client) menu() bool {
	var flag int

	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println(">>>>请输入合法范围内的数字<<<<")
		return false
	}

}

// 查询在线用户
func (client *Client) SelectUsers() {
	sendMsg := "who"
	_, err := client.Send(sendMsg)
	if err != nil {
		fmt.Println("conn Write err:", err)
		return
	}
}

// 私聊模式
func (client *Client) PrivateChat() {
	var remoteName string
	var chatMsg string

	client.SelectUsers()
	fmt.Println(">>>>请输入聊天对象[用户名], exit退出:")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println(">>>>请输入消息内容, exit退出:")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			//消息不为空则发送
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg
				_, err := client.Send(sendMsg)
				if err != nil {
					fmt.Println("conn Write err:", err)
					break
				}
			}

			chatMsg = ""
			fmt.Println(">>>>请输入消息内容, exit退出:")
			fmt.Scanln(&chatMsg)
		}

		client.SelectUsers()
		fmt.Println(">>>>请输入聊天对象[用户名], exit退出:")
		fmt.Scanln(&remoteName)
	}
}

func (client *Client) PublicChat() {
	//提示用户输入消息
	var chatMsg string

	fmt.Println(">>>>请输入聊天内容，exit退出.")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		//发给服务器

		//消息不为空则发送
		if len(chatMsg) != 0 {
			sendMsg := chatMsg
			_, err := client.Send(sendMsg)
			if err != nil {
				fmt.Println("conn Write err:", err)
				break
			}
		}

		chatMsg = ""
		fmt.Println(">>>>请输入聊天内容，exit退出.")
		fmt.Scanln(&chatMsg)
	}

}

func (client *Client) UpdateName() bool {

	fmt.Println(">>>>请输入用户名:")
	fmt.Scanln(&client.Name)

	sendMsg := "rename|" + client.Name
	_, err := client.Send(sendMsg)
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}

	return true
}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {
		}

		//根据不同的模式处理不同的业务
		switch client.flag {
		case 1:
			//公聊模式
			client.PublicChat()
			break
		case 2:
			//私聊模式
			client.PrivateChat()
			break
		case 3:
			//更新用户名
			client.UpdateName()
			break
		}
	}
}

func (client *Client) Send(msg string) (int, error) {
	client.writeMu.Lock()
	defer client.writeMu.Unlock()
	if err := client.conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
		return 0, err
	}
	return len(msg), nil
}

func (client *Client) StartHeartbeat(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		if _, err := client.Send("PING"); err != nil {
			return
		}
	}
}

