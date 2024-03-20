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
	var buf = make([]byte, 512)
	n, err := comm.UdpWriteAndRead(address, 5*time.Second, buf)
	if err != nil {
		log.Fatal(err)
	}

	updAddr := &pb.UDPAddr{}
	err = proto.Unmarshal(buf[0:n], updAddr)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Server response:", updAddr)
}
