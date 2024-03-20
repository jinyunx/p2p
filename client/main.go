package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/jinyunx/p2p/client/comm"
	pb "github.com/jinyunx/p2p/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"os"
	"strconv"
	"time"
)

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	if len(os.Args) != 4 {
		log.Fatalf("usage:%s ip name lport", os.Args[0])
	}
	ip := os.Args[1]
	name := os.Args[2]
	lport, err := strconv.Atoi(os.Args[3])
	if err != nil {
		log.Fatal(err)
	}

	address := fmt.Sprintf("%s:%d", ip, pb.ServerInfo_ServerInfo_Port)

	var updAddr pb.UDPAddr
	getExternalUdp(address, lport, &updAddr)

	updateNode(address, name, &updAddr)

	sendToPeer(address, name, lport)
}

func sendToPeer(address string, name string, lport int) {
	for {
		var target *pb.NodeInfo = nil

		nodeInfo := getNodeInfo(address)
		for _, node := range nodeInfo {
			if node.Name != name {
				target = node
				break
			}
		}
		if target == nil {
			log.Println("no peer found")
			time.Sleep(time.Second)
			continue
		}

		peerAddr := fmt.Sprintf("%s:%d", target.UdpAddr.Ip, target.UdpAddr.Port)
		message := []byte(fmt.Sprintf("hello %s, my name is %s", target.Name, name))
		var buf = make([]byte, 512)

		n, err := comm.UdpWriteAndRead(peerAddr, lport, 5*time.Second, message, buf)
		if err != nil {
			log.Println(err)
		} else {
			log.Println("response:", buf[:n])
		}
		time.Sleep(5 * time.Second)
	}
}

func getNodeInfo(address string) []*pb.NodeInfo {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewP2PClient(conn)

	// Contact the server and print out its response.
	r, err := c.GetNodeInfo(context.Background(), &pb.GetNodeInfoReq{})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Response: %s", r.String())
	return r.GetNodeInfo()
}

func updateNode(address string, name string, updAddr *pb.UDPAddr) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewP2PClient(conn)

	nodeInfo := &pb.NodeInfo{
		Name:    name,
		UdpAddr: updAddr,
	}
	// Contact the server and print out its response.
	r, err := c.UpdateNode(context.Background(), &pb.UpdateNodeReq{
		NodeInfo: nodeInfo,
	})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Response: %s", r.String())
}

func getExternalUdp(address string, lport int, updAddr *pb.UDPAddr) {
	var buf = make([]byte, 512)
	message := []byte("Hello UDP server!")
	n, err := comm.UdpWriteAndRead(address, lport, 5*time.Second, message, buf)
	if err != nil {
		log.Fatal(err)
	}

	err = proto.Unmarshal(buf[0:n], updAddr)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Server response:", updAddr)
}
