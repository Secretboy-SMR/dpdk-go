package engine

import (
	"fmt"

	"dpdk-go/protocol"
)

type TcpEngine struct {
}

func NewTcpEngine() (r *TcpEngine) {
	r = new(TcpEngine)
	return r
}

func (t *TcpEngine) Rx() {
}

func (t *TcpEngine) TxSyn(tcpSrcPort uint16, tcpDstPort uint16, ipv4DstAddr []byte) {
	tcpSynPkt, err := protocol.BuildTcpSynPkt(tcpSrcPort, tcpDstPort, LOCAL_IP_ADDR, ipv4DstAddr, 0)
	if err != nil {
		fmt.Printf("build tcp syn packet error: %v\n", err)
		return
	}
	IPV4_ENGINE.Tx(tcpSynPkt, protocol.IPH_PROTO_TCP, ipv4DstAddr)
}

func (t *TcpEngine) TxSynAck(tcpSrcPort uint16, tcpDstPort uint16, ipv4DstAddr []byte) {
	tcpSynAckPkt, err := protocol.BuildTcpSynAckPkt(tcpSrcPort, tcpDstPort, LOCAL_IP_ADDR, ipv4DstAddr, 0, 0)
	if err != nil {
		fmt.Printf("build tcp syn ack packet error: %v\n", err)
		return
	}
	IPV4_ENGINE.Tx(tcpSynAckPkt, protocol.IPH_PROTO_TCP, ipv4DstAddr)
}

func (t *TcpEngine) TxAck(tcpSrcPort uint16, tcpDstPort uint16, ipv4DstAddr []byte) {
	tcpAckPkt, err := protocol.BuildTcpAckPkt(tcpSrcPort, tcpDstPort, LOCAL_IP_ADDR, ipv4DstAddr, 0, 0)
	if err != nil {
		fmt.Printf("build tcp ack packet error: %v\n", err)
		return
	}
	IPV4_ENGINE.Tx(tcpAckPkt, protocol.IPH_PROTO_TCP, ipv4DstAddr)
}
