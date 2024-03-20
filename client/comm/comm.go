package comm

import (
	"log"
	"net"
	"strconv"
	"time"
)

func UdpWriteAndRead(address string, lport int, timeout time.Duration, buf []byte) (int, error) {
	udpAddr, err := net.ResolveUDPAddr("udp4", address)
	if err != nil {
		log.Println("Invalid server address:", err)
		return 0, err
	}

	ludpAddr, err := net.ResolveUDPAddr("udp4", ":"+strconv.Itoa(lport))
	if err != nil {
		log.Println("Invalid server address:", err)
		return 0, err
	}

	// 创建UDP连接
	conn, err := net.DialUDP("udp", ludpAddr, udpAddr)
	if err != nil {
		log.Println("Error connecting to UDP server:", err)
		return 0, err
	}
	defer conn.Close()

	// 发送消息到服务器
	message := []byte("Hello UDP server!")
	_, err = conn.Write(message)
	if err != nil {
		log.Println("Error sending message:", err)
		return 0, err
	}

	// 设置读取超时
	err = conn.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		log.Println("Error setting read deadline:", err)
		return 0, err
	}

	// 读取服务器响应
	n, err := conn.Read(buf[0:])
	if err != nil {
		if e, ok := err.(net.Error); ok && e.Timeout() {
			log.Println("Read timeout:", err)
		} else {
			log.Println("Error reading response:", err)
		}
		return 0, err
	}
	return n, nil
}
