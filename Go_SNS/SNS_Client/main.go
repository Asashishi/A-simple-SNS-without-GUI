package main

import (
	"SNS_Client/client"
	"flag"
	"fmt"
	"github.com/google/uuid"
	"net"
	"runtime"
	"time"
)

var serverIP string
var serverPort int
var Name string

// 进程初始化-绑定命令行参数 -ip 8.218.247.195 -port 5195 -name Asashishi
func init() {
	newUUID, err := uuid.NewRandom()
	if err != nil {
		fmt.Println("Error generating UUID: ", err)
	}

	flag.StringVar(&serverIP, "ip", "127.0.0.1", "Set server ip address (Default: 127.0.0.1)")
	flag.IntVar(&serverPort, "port", 5195, "Set server port (Default: 5195)")
	flag.StringVar(&Name, "name", newUUID.String(), "Set server name (Default: GoLand)")
}

// 心跳链接
func hart_bit(conn net.Conn) {
	for {
		conn.Write([]byte("HART"))
		time.Sleep(15 * time.Second)
	}
}

// 手动触发内存回收
func timeToGC() {
	for {
		runtime.GC()
		time.Sleep(15 * time.Second)
	}
}

func main() {

	// 解析命令行
	flag.Parse()

	// 打印信息
	fmt.Printf(
		"serverIP: %v\n"+
			"serverPort: %v\n"+
			"ClientUserName: %v\n",
		serverIP, serverPort, Name,
	)

	// 链接服务器
	Client := client.NewClient(serverIP, serverPort, Name)
	if Client == nil {
		fmt.Println("Failed to connect SNS_Server!")
		return
	}
	userName := "-rename " + Name
	_, err := Client.Conn.Write([]byte(userName))
	if err != nil {
		fmt.Println("Failed to connect SNS_Server!")
		return
	}

	fmt.Println("Connected to SNS_Server!")

	// 处理服务端返回的消息
	go Client.DealResponse()
	go hart_bit(Client.Conn)

	go timeToGC()

	Client.Run()
}
