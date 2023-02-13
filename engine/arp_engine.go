package engine

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/FlourishingWorld/dpdk-go/dpdk"
	"github.com/FlourishingWorld/dpdk-go/protocol"
)

type ArpEngine struct {
	// arp缓存表 key:ip value:mac
	ArpCacheTable     map[uint32]uint64
	ArpCacheTableLock sync.RWMutex
}

func NewArpEngine() (r *ArpEngine) {
	r = new(ArpEngine)
	r.ArpCacheTable = make(map[uint32]uint64)
	return r
}

func (a *ArpEngine) ConvMacAddrToUint64(macAddr []byte) (macAddrUint64 uint64) {
	macAddrUint64 = uint64(0)
	macAddrUint64 += uint64(macAddr[0]) << 40
	macAddrUint64 += uint64(macAddr[1]) << 32
	macAddrUint64 += uint64(macAddr[2]) << 24
	macAddrUint64 += uint64(macAddr[3]) << 16
	macAddrUint64 += uint64(macAddr[4]) << 8
	macAddrUint64 += uint64(macAddr[5]) << 0
	return macAddrUint64
}

func (a *ArpEngine) ConvUint64ToMacAddr(macAddrUint64 uint64) (macAddr []byte) {
	macAddr = make([]byte, 6)
	macAddr[0] = uint8(macAddrUint64 >> 40)
	macAddr[1] = uint8(macAddrUint64 >> 32)
	macAddr[2] = uint8(macAddrUint64 >> 24)
	macAddr[3] = uint8(macAddrUint64 >> 16)
	macAddr[4] = uint8(macAddrUint64 >> 8)
	macAddr[5] = uint8(macAddrUint64 >> 0)
	return macAddr
}

func (a *ArpEngine) GetArpCache(ipAddr []byte) (macAddr []byte) {
	ipAddrUint32 := protocol.ConvIpAddrToUint32(ipAddr)
	a.ArpCacheTableLock.RLock()
	macAddrUint64, exist := a.ArpCacheTable[ipAddrUint32]
	a.ArpCacheTableLock.RUnlock()
	if !exist {
		// 不存在则发起ARP询问并返回空
		arpPkt, err := protocol.BuildArpPkt(protocol.ARP_REQUEST, LOCAL_MAC_ADDR, LOCAL_IP_ADDR, BROADCAST_MAC_ADDR, ipAddr)
		if err != nil {
			fmt.Printf("build arp packet error: %v\n", err)
			return
		}
		ethFrm, err := protocol.BuildEthFrm(arpPkt, BROADCAST_MAC_ADDR, LOCAL_MAC_ADDR, protocol.ETH_PROTO_ARP)
		if err != nil {
			fmt.Printf("build ethernet frame error: %v\n", err)
			return
		}
		if DEBUG {
			fmt.Printf("tx arp pkt, eth frm len: %v, eth frm data: %v\n", len(ethFrm), ethFrm)
		}
		dpdk.DPDK_TX_CHAN <- ethFrm
		return nil
	}
	macAddr = a.ConvUint64ToMacAddr(macAddrUint64)
	return macAddr
}

func (a *ArpEngine) SetArpCache(ipAddr []byte, macAddr []byte) {
	ipAddrUint32 := protocol.ConvIpAddrToUint32(ipAddr)
	macAddrUint64 := a.ConvMacAddrToUint64(macAddr)
	a.ArpCacheTableLock.Lock()
	a.ArpCacheTable[ipAddrUint32] = macAddrUint64
	a.ArpCacheTableLock.Unlock()
}

func (a *ArpEngine) Handle(ethPayload []byte, ethSrcMac []byte) {
	arpOption, arpSrcMac, arpSrcAddr, _, arpDstAddr, err := protocol.ParseArpPkt(ethPayload)
	if err != nil {
		fmt.Printf("parse arp packet error: %v\n", err)
		return
	}
	if !bytes.Equal(arpSrcMac, ethSrcMac) {
		fmt.Printf("arp packet src mac addr not match\n")
		return
	}
	a.SetArpCache(arpSrcAddr, arpSrcMac)
	// 对目的IP为本机的ARP询问请求进行回应
	if arpOption == protocol.ARP_REQUEST && bytes.Equal(arpDstAddr, LOCAL_IP_ADDR) {
		arpPkt, err := protocol.BuildArpPkt(protocol.ARP_REPLY, LOCAL_MAC_ADDR, LOCAL_IP_ADDR, arpSrcMac, arpSrcAddr)
		if err != nil {
			fmt.Printf("build arp packet error: %v\n", err)
			return
		}
		ethFrm, err := protocol.BuildEthFrm(arpPkt, arpSrcMac, LOCAL_MAC_ADDR, protocol.ETH_PROTO_ARP)
		if err != nil {
			fmt.Printf("build ethernet frame error: %v\n", err)
			return
		}
		if DEBUG {
			fmt.Printf("tx arp pkt, eth frm len: %v, eth frm data: %v\n", len(ethFrm), ethFrm)
		}
		dpdk.DPDK_TX_CHAN <- ethFrm
	}
}
