package client

import (
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

type Client struct {
	ServerIP   string
	ServerPort int
	Name       string
	Conn       net.Conn
	flag       int
}

// NewClient 创建链接
func NewClient(serverIP string, serverPort int, Name string) *Client {
	// 创建客户端对象
	client := &Client{
		ServerIP:   serverIP,
		ServerPort: serverPort,
		Name:       Name,
		flag:       1072903224,
	}
	// 链接Server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", client.ServerIP, client.ServerPort))
	if err != nil {
		fmt.Println("net.Dial error", err)
		return nil
	}
	client.Conn = conn
	return client
}

// ClientMenu 客户端菜单
func (client *Client) clientMenu() bool {
	var flag int
	time.Sleep(time.Second)
	fmt.Println(
		"ClientMenu:\n" +
			"1: Public_SNS_Mode\n" +
			"2: Private_SNS_Mode\n" +
			"3: UserReName\n" +
			"4: LogOut",
	)
	_, err := fmt.Scanln(&flag)
	if err != nil {
		return false
	}
	if flag >= 1 && flag <= 4 {
		client.flag = flag
		return true
	} else {
		fmt.Println("Please enter a valid flag!")
		return false
	}
}

// 公聊
func (client *Client) PublicChat() {
	var chartMsg string
	for {
		// 提示用户输入信息
		fmt.Println("Please input your message (input [exit] to exit public chart): ")
		fmt.Scanln(&chartMsg)
		if chartMsg != "exit" {
			if len(chartMsg) != 0 {
				_, err := client.Conn.Write([]byte(chartMsg))
				if err != nil {
					fmt.Println("Failed to send your message:", err)
					break
				}
			}
		} else {
			break
		}
	}
}

// 查询在线用户
func (client *Client) SelectOnlineUser() {
	_, err := client.Conn.Write([]byte("-ou" + "\n"))
	if err != nil {
		fmt.Println("Failed to send your message:", err)
	}
}

// 私聊
func (client *Client) PrivateChat() {
	var remoteUser string
	var chartMsg string
	for {
		client.SelectOnlineUser()
		fmt.Println("Please enter a valid user (input [exit] to exit private chart): ")
		fmt.Scanln(&remoteUser)
		if remoteUser != "exit" {
			for {
				fmt.Scanln(&chartMsg)
				fmt.Println("Please enter a message:")
				if chartMsg != "exit" {
					if len(chartMsg) != 0 {
						_, err := client.Conn.Write([]byte("-private_to " + remoteUser + " " + chartMsg + "\n"))
						if err != nil {
							fmt.Println("Failed to send your message:", err)
							break
						}
					}
				} else {
					break
				}
			}
		} else {
			break
		}
	}

}

// 用户改名
func (client *Client) userReName() bool {
	fmt.Print("Enter your new name: ")
	_, err0 := fmt.Scanln(&client.Name)
	if err0 != nil {
		fmt.Println("Failed to get your input: ", err0)
		return false
	}
	sendMsg := "-rename " + client.Name + "\n"
	_, err1 := client.Conn.Write([]byte(sendMsg))
	if err1 != nil {
		fmt.Println("Failed to set your new name: ", err1)
		return false
	} else {
		return true
	}
}

// 端开链接
func (client *Client) logOut() {
	_, err := client.Conn.Write([]byte("-offline" + "\n"))
	if err != nil {
		fmt.Println("Failed to send your message:", err)
	}
	client.Conn.Close()
	fmt.Println("you have been logout!")
	os.Exit(0)
}

// DealResponse 处理服务器返回消息方法
func (client *Client) DealResponse() {
	// 如果有数据则拷贝数据到标准输出,永久监听
	_, err := io.Copy(os.Stdout, client.Conn)
	if err != nil {
		fmt.Println("io.Copy error", err)
	}
	// 等价于
	//for {
	//	buf := make([]byte, 4096)
	//	_, err := client.Conn.Read(buf)
	//	if err != nil {
	//		fmt.Println("Read buf error", err)
	//	}
	//	fmt.Println(string(buf))
	//}
}

// Run 方法
func (client *Client) Run() {
	for client.flag != 0 {
		for client.clientMenu() != true {
		}
		// 处理菜单
		switch client.flag {
		case 1:
			// Public
			client.PublicChat()
		case 2:
			// Private
			client.PrivateChat()
		case 3:
			client.userReName()
		case 4:
			// Logout
			client.logOut()

		}
	}
}
