package main

import (
	"encoding/json"
	"fmt"
	pb "github.com/jinyunx/p2p/proto"
	"log"
	"net"
	"os"
)

func udpServer(port string) {
	udpAddr, err := net.ResolveUDPAddr("udp4", port)
	if err != nil {
		fmt.Println("Invalid address:", err)
		os.Exit(1)
	}

	// 创建UDP监听
	log.Println("Listen udp", port)
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("Error listening on UDP port:", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Println("UDP server listening on port", port)

	// 无限循环，等待并处理数据
	for {
		handleClient(conn)
	}
}

func handleClient(conn *net.UDPConn) {
	var buf [512]byte

	// 读取数据
	n, addr, err := conn.ReadFromUDP(buf[0:])
	if err != nil {
		log.Println(err)
		return
	}

	// 打印接收到的消息
	fmt.Println("Received ", string(buf[0:n]), " from ", addr)

	udpAddr := pb.UDPAddr{
		Ip:   string(addr.IP),
		Port: int32(addr.Port),
		Zone: addr.Zone,
	}

	addrJson, err := json.Marshal(udpAddr)
	if err != nil {
		log.Println(err)
	}

	// 发送响应
	_, err = conn.WriteToUDP(addrJson, addr)
	if err != nil {
		fmt.Println("Error sending response:", err)
	}
}
