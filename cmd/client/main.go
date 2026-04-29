package main

import (
	"flag"
	"fmt"
	"time"

	"IM-System/internal/client"
)

var serverIp string
var serverPort int
var wsPath string
var heartbeatIntervalSeconds int

// ./client -ip 127.0.0.1 -port 8888
func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址(默认是127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口(默认是8888)")
	flag.StringVar(&wsPath, "path", "/ws", "设置WebSocket路径(默认是/ws)")
	flag.IntVar(&heartbeatIntervalSeconds, "hb", 300, "设置心跳间隔秒数(默认是300秒)")
}

func main() {
	//命令行解析
	flag.Parse()

	cli := client.NewClient(serverIp, serverPort, wsPath)
	if cli == nil {
		fmt.Println(">>>>> 链接服务器失败...")
		return
	}

	//单独开启一个goroutine去处理server的回执消息
	go cli.DealResponse()
	go cli.StartHeartbeat(time.Duration(heartbeatIntervalSeconds) * time.Second)

	fmt.Println(">>>>>链接服务器成功...")

	//启动客户端的业务
	cli.Run()
}
