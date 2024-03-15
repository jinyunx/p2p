package main

import (
	"fmt"
	pb "github.com/jinyunx/p2p/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"os"
)

func main() {

	if len(os.Args) != 2 {
		log.Fatalf("usage:%s ip", os.Args[0])
	}
	ip := os.Args[1]

	address := fmt.Sprintf("%s:%d", ip, pb.ServerInfo_ServerInfo_Port)
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
