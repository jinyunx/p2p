package stun

import (
	"github.com/jinyunx/p2p/client/comm"
	"log"
	"time"
)

func BindingRequest(addr string) error {
	stunMsg, err := InitStunMsg(StunMsgType_BindingRequest, []Attr{})
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println(stunMsg)

	bin, err := stunMsg.Marshal()
	if err != nil {
		log.Println(err)
		return err
	}

	var resp [1024]byte
	n, err := comm.UdpWriteAndRead(addr, 12345, time.Second, bin, resp[:])
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println(n, resp[:n])

	var respStunMsg StunMsg
	err = respStunMsg.UnMarshal(resp[:n])
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println(respStunMsg.String())
	return nil
}
