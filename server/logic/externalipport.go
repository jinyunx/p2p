package logic

import (
	"fmt"
	pb "github.com/jinyunx/p2p/proto"
	"google.golang.org/grpc/peer"
)
import "golang.org/x/net/context"

func GetExternalIpPort(ctx context.Context, in *pb.GetExternalIpPortReq) (*pb.GetExternalIpPortResp, error) {
	// 获取对端信息
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to get peer from context")
	}

	// p.Addr是net.Addr类型，包含了IP地址和端口
	fmt.Printf("Client IP address: %s\n", p.Addr.String())
	return &pb.GetExternalIpPortResp{
		Addr:    p.Addr.String(),
		Network: p.Addr.Network(),
	}, nil
}
