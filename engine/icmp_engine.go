package engine

import (
	"fmt"
	"time"

	"github.com/FlourishingWorld/dpdk-go/protocol"
)

type IcmpEngine struct {
}

func NewIcmpEngine() (r *IcmpEngine) {
	r = new(IcmpEngine)
	return r
}

func (i *IcmpEngine) NetworkStateCheck() {
	ticker := time.NewTicker(time.Second * 10)
	seq := uint16(0)
	for {
		<-ticker.C
		seq++
		icmpPkt, err := protocol.BuildIcmpPkt(protocol.ICMP_DEFAULT_PAYLOAD, protocol.ICMP_REQUEST, []byte{0x00, 0x01}, []byte{uint8(seq >> 8), uint8(seq)})
		if err != nil {
			fmt.Printf("build icmp packet error: %v\n", err)
			continue
		}
		IPV4_ENGINE.Tx(icmpPkt, protocol.IPH_PROTO_ICMP, GATEWAY_IP_ADDR)
	}
}

func (i *IcmpEngine) Handle(ipv4Payload []byte, ipv4SrcAddr []byte) {
	icmpPayload, icmpType, icmpId, icmpSeq, err := protocol.ParseIcmpPkt(ipv4Payload)
	if err != nil {
		fmt.Printf("parse icmp packet error: %v\n", err)
		return
	}
	if icmpType == protocol.ICMP_REQUEST {
		// 构造ICMP响应包
		icmpPkt, err := protocol.BuildIcmpPkt(icmpPayload, protocol.ICMP_REPLY, icmpId, icmpSeq)
		if err != nil {
			fmt.Printf("build icmp packet error: %v\n", err)
			return
		}
		IPV4_ENGINE.Tx(icmpPkt, protocol.IPH_PROTO_ICMP, ipv4SrcAddr)
	}
}
