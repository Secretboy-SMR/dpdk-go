package engine

import (
	"fmt"

	"dpdk-go/protocol"
	"dpdk-go/protocol/kcp"
)

type UdpEngine struct {
}

func NewUdpEngine() (r *UdpEngine) {
	r = new(UdpEngine)
	kcp.UdpTx = r.Tx
	return r
}

func (u *UdpEngine) Rx(ipv4Payload []byte, ipv4SrcAddr []byte) {
	udpPayload, udpSrcPort, udpDstPort, err := protocol.ParseUdpPkt(ipv4Payload, ipv4SrcAddr, LOCAL_IP_ADDR)
	if err != nil {
		fmt.Printf("parse udp packet error: %v\n", err)
		return
	}
	kcp.UdpRx(udpPayload, udpSrcPort, udpDstPort, ipv4SrcAddr)
}

func (u *UdpEngine) Tx(udpPayload []byte, udpSrcPort uint16, udpDstPort uint16, ipv4DstAddr []byte) {
	udpPkt, err := protocol.BuildUdpPkt(udpPayload, udpSrcPort, udpDstPort, LOCAL_IP_ADDR, ipv4DstAddr)
	if err != nil {
		fmt.Printf("build udp packet error: %v\n", err)
		return
	}
	IPV4_ENGINE.Tx(udpPkt, protocol.IPH_PROTO_UDP, ipv4DstAddr)
}
