package logic

import (
	pb "github.com/jinyunx/p2p/proto"
	"log"
	"sync"
)
import "golang.org/x/net/context"

type NodesMap struct {
	mu    sync.Mutex
	nodes map[string]pb.NodeInfo
}

var nodeInfo = NodesMap{nodes: make(map[string]pb.NodeInfo)}

func UpdateNode(ctx context.Context, in *pb.UpdateNodeReq) (*pb.UpdateNodeResp, error) {
	log.Println("UpdateNode req", in)
	nodeInfo.mu.Lock()
	nodeInfo.nodes[in.GetNodeInfo().GetName()] = *in.GetNodeInfo()
	nodeInfo.mu.Unlock()
	return &pb.UpdateNodeResp{}, nil
}

func GetNodeInfo(ctx context.Context, in *pb.GetNodeInfoReq) (*pb.GetNodeInfoResp, error) {
	log.Println("GetNodeInfo req", in)
	var out = &pb.GetNodeInfoResp{}
	nodeInfo.mu.Lock()
	nodesMap := nodeInfo.nodes
	nodeInfo.mu.Unlock()

	n := out.GetNodeInfo()
	for k, _ := range nodesMap {
		n = append(n, &pb.NodeInfo{
			Name:    nodesMap[k].Name,
			UdpAddr: nodesMap[k].UdpAddr,
		})
	}
	return out, nil
}
