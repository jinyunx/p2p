package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	pb "github.com/jinyunx/p2p/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"time"
)

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	if len(os.Args) != 2 {
		log.Fatalf("usage:%s ip", os.Args[0])
	}
	ip := os.Args[1]

	address := fmt.Sprintf("%s:%d", ip, pb.ServerInfo_ServerInfo_Port)
	rpcRun(address)
	udpRun(address)
}

func rpcRun(address string) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewP2PClient(conn)

	// Contact the server and print out its response.
	r, err := c.GetExternalIpPort(context.Background(), &pb.GetExternalIpPortReq{})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Response: %s", r.String())
}

func udpRun(address string) {
	udpAddr, err := net.ResolveUDPAddr("udp4", address)
	if err != nil {
		log.Println("Invalid server address:", err)
		os.Exit(1)
	}

	// 创建UDP连接
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Println("Error connecting to UDP server:", err)
		os.Exit(1)
	}
	defer conn.Close()

	// 发送消息到服务器
	message := []byte("Hello UDP server!")
	_, err = conn.Write(message)
	if err != nil {
		log.Println("Error sending message:", err)
		os.Exit(1)
	}

	// 设置读取超时
	err = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		log.Println("Error setting read deadline:", err)
		os.Exit(1)
	}

	// 读取服务器响应
	var buf [512]byte
	n, err := conn.Read(buf[0:])
	if err != nil {
		if e, ok := err.(net.Error); ok && e.Timeout() {
			log.Println("Read timeout:", err)
		} else {
			log.Println("Error reading response:", err)
		}
		os.Exit(1)
	}

	// 打印服务器响应
	updAddr := &pb.UDPAddr{}
	err = proto.Unmarshal(buf[0:n], updAddr)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Server response:", updAddr)
}
