package stun

import (
	"github.com/jinyunx/p2p/client/comm"
	"log"
	"time"
)

/*
   The flow makes use of three tests.  In test I, the client sends a
   STUN Binding Request to a server, without any flags set in the
   CHANGE-REQUEST attribute, and without the RESPONSE-ADDRESS attribute.
   This causes the server to send the response back to the address and
   port that the request came from.  In test II, the client sends a
   Binding Request with both the "change IP" and "change port" flags
   from the CHANGE-REQUEST attribute set.  In test III, the client sends
   a Binding Request with only the "change port" flag set.

+--------+
                        |  Test  |
                        |   I    |
                        +--------+
                             |
                             |
                             V
                            /\              /\
                         N /  \ Y          /  \ Y             +--------+
          UDP     <-------/Resp\--------->/ IP \------------->|  Test  |
          Blocked         \ ?  /          \Same/              |   II   |
                           \  /            \? /               +--------+
                            \/              \/                    |
                                             | N                  |
                                             |                    V
                                             V                    /\
                                         +--------+  Sym.      N /  \
                                         |  Test  |  UDP    <---/Resp\
                                         |   II   |  Firewall   \ ?  /
                                         +--------+              \  /
                                             |                    \/
                                             V                     |Y
                  /\                         /\                    |
   Symmetric  N  /  \       +--------+   N  /  \                   V
      NAT  <--- / IP \<-----|  Test  |<--- /Resp\               Open
                \Same/      |   I    |     \ ?  /               Internet
                 \? /       +--------+      \  /
                  \/                         \/
                  |                           |Y
                  |                           |
                  |                           V
                  |                           Full
                  |                           Cone
                  V              /\
              +--------+        /  \ Y
              |  Test  |------>/Resp\---->Restricted
              |   III  |       \ ?  /
              +--------+        \  /
                                 \/
                                  |N
                                  |       Port
                                  +------>Restricted
*/

func BindingRequest(addr string) error {
	var attrs []Attr
	var changeRequest ChangeRequest
	changeRequest.Init(true, true)
	attrs = append(attrs, &changeRequest)
	stunMsg, err := InitStunMsg(StunMsgType_BindingRequest, attrs)
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
	log.Printf("%x\n", bin)

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
	log.Println(respStunMsg.Attrs)
	return nil
}
