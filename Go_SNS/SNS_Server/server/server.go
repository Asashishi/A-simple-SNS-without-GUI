package server

import (
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

type Server struct {
	IP   string
	Port int

	// 在线用户列表 指向User结构体
	OnlineMap map[string]*User
	Maplock   sync.RWMutex

	// 消息广播管道
	Message chan string
}

// NewServer Server接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		// 接收IP和端口属性
		IP:   ip,
		Port: port,

		// 创建一个OnlineMap和管道
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// 监听广播消息的协程
func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message
		this.Maplock.Lock()
		// 遍历管道中的信息 发送给User
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.Maplock.Unlock()
	}
}

// 广播消息的方法
func (this *Server) BroadCast(user *User, msg string) {
	// 拼接消息
	sendMsg := "[" + user.Addr + "] " + user.Name + ": " + msg + " (public)\n"
	// 放入管道
	this.Message <- sendMsg
}

// 处理客户端连接
func (this *Server) Handler(conn net.Conn) {

	fmt.Printf("Conn succeeded from: %v\n", conn.RemoteAddr().String())

	user := NewUser(this, conn)

	// 将用户加入到OnlineMap表中,并向其他用户广播
	user.Online()

	// 监听用户是否活跃的管道
	isAlive := make(chan bool)

	// 接收客户端发送的消息并广播
	go func() {
		for {
			buf := make([]byte, 4096)
			n, err := conn.Read(buf)
			if err != nil && err != io.EOF {
				if strings.Contains(err.Error(), "An existing connection was forcibly closed by the remote host.") {
					user.Offline()
					fmt.Printf("User %v has logout!\n", user.Name)
					return
				}
				fmt.Println("Read user mess error:", err)
			}

			if n == 0 {
				user.Offline()
				fmt.Printf("User %v has logout!\n", user.Name)
				return
			}
			// 用户模块处理消息
			msg := string(buf[:n])
			// 判定用户发送的不为心跳链接时处理消息并刷新用户活跃状态
			if msg != "HART" {
				user.DoMessage(msg)
				// 判定用户活跃
				isAlive <- true
			}
		}
	}() // 调用匿名函数

	// 当前handeler阻塞
	for {
		select {
		// 当前用户活跃则重置定时器 执行case之后触发下一份case的条件,但未满足下一个case
		case <-isAlive:
		// 超时则触发case
		case <-time.After(time.Second * 180):
			// 超时处理
			user.SendMsg("Due to prolonged inactivity, you have been forcibly logged out, Please reconnect!\n")
			conn.Close()
			user.Offline()
		}
	}
}

// Start 启动服务器方法
func (this *Server) Start() {

	fmt.Printf("Go SNS_Server starting on %s:%d\n", this.IP, this.Port)

	// 启动监听msg的协程
	go this.ListenMessager()

	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.IP, this.Port))
	if err != nil {
		fmt.Printf("Listener error: %v\n", err)
	}
	// close listen socket
	defer listener.Close()

	for {

		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Listener accept error: %v\n", err)
			continue
		}

		// do handler
		go this.Handler(conn)
	}

}
