package engine

import (
	"bytes"
	"fmt"

	"dpdk-go/dpdk"
	"dpdk-go/protocol"
)

type Ipv4Engine struct {
}

func NewIpv4Engine() (r *Ipv4Engine) {
	r = new(Ipv4Engine)
	return r
}

func (i *Ipv4Engine) Handle(ethPayload []byte) {
	ipv4Payload, ipv4HeadProto, ipv4SrcAddr, ipv4DstAddr, err := protocol.ParseIpv4Pkt(ethPayload)
	if err != nil {
		fmt.Printf("parse ip packet error: %v\n", err)
		return
	}
	if !bytes.Equal(ipv4DstAddr, LOCAL_IP_ADDR) {
		return
	}
	switch ipv4HeadProto {
	case protocol.IPH_PROTO_ICMP:
		ICMP_ENGINE.Handle(ipv4Payload, ipv4SrcAddr)
	case protocol.IPH_PROTO_UDP:
		UDP_ENGINE.Rx(ipv4Payload, ipv4SrcAddr)
	case protocol.IPH_PROTO_TCP:
		TCP_ENGINE.Rx()
	default:
	}
}

func (i *Ipv4Engine) Tx(ipv4Payload []byte, ipv4HeadProto uint8, ipv4DstAddr []byte) {
	ipv4Pkt, err := protocol.BuildIpv4Pkt(ipv4Payload, ipv4HeadProto, LOCAL_IP_ADDR, ipv4DstAddr)
	if err != nil {
		fmt.Printf("build ip packet error: %v\n", err)
		return
	}
	// ip路由
	var ethDstMac []byte = nil
	localIpUint32 := protocol.ConvIpAddrToUint32(LOCAL_IP_ADDR)
	dstIpUint32 := protocol.ConvIpAddrToUint32(ipv4DstAddr)
	networkMaskUint32 := protocol.ConvIpAddrToUint32(NETWORK_MASK)
	if localIpUint32&networkMaskUint32 == dstIpUint32&networkMaskUint32 {
		// 同一子网
		ethDstMac = ARP_ENGINE.GetArpCache(ipv4DstAddr)
	} else {
		// 不同子网
		ethDstMac = ARP_ENGINE.GetArpCache(GATEWAY_IP_ADDR)
	}
	ethFrm, err := protocol.BuildEthFrm(ipv4Pkt, ethDstMac, LOCAL_MAC_ADDR, protocol.ETH_PROTO_IP)
	if err != nil {
		fmt.Printf("build ethernet frame error: %v\n", err)
		return
	}
	fmt.Printf("tx ip pkt, eth frm len: %v, eth frm data: %v\n", len(ethFrm), ethFrm)
	dpdk.DPDK_TX_CHAN <- ethFrm
}
