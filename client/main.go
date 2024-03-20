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

	var nodeInfo []*pb.NodeInfo
	getNodeInfo(address, nodeInfo)
}

func getNodeInfo(address string, nodeInfo []*pb.NodeInfo) {
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
	nodeInfo = r.GetNodeInfo()
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
	n, err := comm.UdpWriteAndRead(address, lport, 5*time.Second, buf)
	if err != nil {
		log.Fatal(err)
	}

	err = proto.Unmarshal(buf[0:n], updAddr)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Server response:", updAddr)
}
