package main

import (
	"fmt"
	pb "github.com/jinyunx/p2p/proto"
	"github.com/jinyunx/p2p/server/logic"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
)

type server struct {
	pb.UnimplementedP2PServer
}

func (s *server) GetExternalIpPort(ctx context.Context, in *pb.GetExternalIpPortReq) (*pb.GetExternalIpPortResp, error) {
	log.Println("GetExternalIpPort req", in)
	return logic.GetExternalIpPort(ctx, in)
}

func main() {
	port := fmt.Sprintf(":%d", pb.ServerInfo_ServerInfo_Port)
	go udpServer(port)

	log.Println("Listen tcp rpc", port)
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterP2PServer(s, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
