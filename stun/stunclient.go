package stun

import (
	"github.com/jinyunx/p2p/client/comm"
	"log"
	"time"
)

func BindingRequest(addr string) error {
	stunMsg, err := InitStunMsg(StunMsgType_BindingRequest, []TLV{})
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println(stunMsg)
	bin := stunMsg.Marshal()
	var resp [1024]byte
	n, err := comm.UdpWriteAndRead(addr, 12345, time.Second, bin, resp[:])
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println(n, resp[:n])

	var respStunMsg StunMsg
	respStunMsg.UnMarshal(resp[:n])
	log.Println(respStunMsg)
	return nil
}
