package public

import (
	"log"
	"net"
	"os"
)

type UdpDataHandler func(*net.UDPConn, []byte, *net.UDPAddr)

func UdpServer(addr string, handle UdpDataHandler) {
	udpAddr, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		log.Println("Invalid address:", err)
		os.Exit(1)
	}

	// 创建UDP监听
	log.Println("Listen udp", addr)
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Println("Error listening on UDP port:", err)
		os.Exit(1)
	}
	defer conn.Close()

	log.Println("UDP server listening on port", addr)

	// 无限循环，等待并处理数据
	for {
		handleClient(conn, handle)
	}
}

func handleClient(conn *net.UDPConn, handle UdpDataHandler) {
	var buf [8 * 1024]byte

	// 读取数据
	n, addr, err := conn.ReadFromUDP(buf[0:])
	if err != nil {
		log.Println(err)
		return
	}

	// 打印接收到的消息
	log.Println("Received from ", addr)
	handle(conn, buf[0:n], addr)
}
