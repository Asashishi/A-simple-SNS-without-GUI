package server

import (
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

// ListenMessage 监听User channel
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		// 判断是否为此用户发送的信息
		if !strings.Contains(msg, this.Name) {
			for i := 0; i < 3; i++ {
				_, err := this.conn.Write([]byte(msg))
				if err != nil && nil != io.EOF {
					fmt.Printf("Flied to send message %v retry %v/3 times\n", err, i+1)
					time.Sleep(time.Second)
				}
				break
			}
		}
	}
}

// NewUser User 构造器
func NewUser(server *Server, conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		server: server,
		conn:   conn,
	}
	go user.ListenMessage()
	return user
}

// Online 用户上线
func (this *User) Online() {

	// 放入Map
	this.server.Maplock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.Maplock.Unlock()

	this.server.BroadCast(this, "was online!")
}

// Offline 用户下线
func (this *User) Offline() {
	// 从map中取出
	this.server.Maplock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.Maplock.Unlock()
	this.conn.Close()

	this.server.BroadCast(this, "was offline!")
}

// SendMsg 当前User对应的客户端接收消息
func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

// 帮助指令
func (this *User) cmdGuide() {
	this.SendMsg(
		"[-ou]: List online users" +
			"\n[-rename]: Send message like [-rename name] to change your name" +
			"\n[-private_to] Send like [-private_to username message] to send private message" +
			"\n[-offline] offline\n",
	)
}

// 在线用户清单
func (this *User) ListOlineUser(msg string) {
	onlineInfo := ""
	this.server.Maplock.Lock()
	for _, user := range this.server.OnlineMap {
		onlineInfo += user.Name + " is online\n"
		this.SendMsg(onlineInfo)
	}
	this.server.Maplock.Unlock()
}

// 修改用户名
func (this *User) ReName(msg string) {
	newName := strings.Split(msg, " ")[1]
	_, used := this.server.OnlineMap[newName]
	if used {
		this.SendMsg("This Name already exists!\n")
	} else {
		this.server.Maplock.Lock()
		delete(this.server.OnlineMap, this.Name)
		this.server.OnlineMap[newName] = this
		this.server.Maplock.Unlock()
		this.Name = newName

		this.SendMsg("You have been renamed to: " + newName + "\n")
	}
}

// 指定用户进行私聊
func (this *User) PrivateMessage(msg string) {
	who := strings.Split(msg, " ")[1]
	if who == "" {
		this.SendMsg("You hava to send a message like [-private_to username message] to use private message function!\n")
		return
	}
	remoteUser, find := this.server.OnlineMap[who]
	if !find {
		this.SendMsg("User do not exist!\n")
		return
	}
	content := strings.Split(msg, " ")[2]
	if content == "" {
		this.SendMsg("Failed to get your message, you hava to send a message like [-private_to username message] to use private message function!\n")
		return
	}
	content = "[" + this.Addr + "] " + this.Name + ": " + content + " (private)\n"
	remoteUser.SendMsg(content)
}

// DoMessage 用户处理消息
func (this *User) DoMessage(msg string) {

	switch {
	case msg == "-help":
		this.cmdGuide()
	case msg == "-ou":
		this.ListOlineUser(msg)
	case strings.Contains(msg, "-rename"):
		this.ReName(msg)
	case strings.Contains(msg, "-private_to"):
		this.PrivateMessage(msg)
	case msg == "-offline":
		this.Offline()
	default:
		this.server.BroadCast(this, msg)
		fmt.Printf("Messege: %v Send by: [%v] %v\n", msg, this.Addr, this.Name)
	}
}
