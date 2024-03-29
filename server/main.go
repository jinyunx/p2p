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

func (s *server) UpdateNode(ctx context.Context, in *pb.UpdateNodeReq) (*pb.UpdateNodeResp, error) {
	log.Println("UpdateNode req", in)
	return logic.UpdateNode(ctx, in)
}

func (s *server) GetNodeInfo(ctx context.Context, in *pb.GetNodeInfoReq) (*pb.GetNodeInfoResp, error) {
	log.Println("GetNodeInfo req", in)
	return logic.GetNodeInfo(ctx, in)
}

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
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
