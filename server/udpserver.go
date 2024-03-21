package main

import (
	"github.com/golang/protobuf/proto"
	pb "github.com/jinyunx/p2p/proto"
	"github.com/jinyunx/p2p/public"
	"log"
	"net"
)

func udpServer(port string) {
	public.UdpServer(port, handleData)
}

func handleData(conn *net.UDPConn, buf []byte, addr *net.UDPAddr) {
	log.Println("Received ", string(buf), " from ", addr)

	udpAddr := &pb.UDPAddr{
		Ip:   addr.IP.String(),
		Port: int32(addr.Port),
		Zone: addr.Zone,
	}

	marshalAddr, err := proto.Marshal(udpAddr)
	if err != nil {
		log.Println(udpAddr, err)
		return
	}

	// 发送响应
	_, err = conn.WriteToUDP(marshalAddr, addr)
	if err != nil {
		log.Println("Error sending response:", err)
	}
}
