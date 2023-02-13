package engine

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/FlourishingWorld/dpdk-go/dpdk"
	"github.com/FlourishingWorld/dpdk-go/protocol"
)

var DEBUG = false

func EnableDebug() {
	DEBUG = true
}

func DisableDebug() {
	DEBUG = false
}

var (
	BROADCAST_MAC_ADDR        = []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	LOCAL_MAC_ADDR     []byte = nil
	LOCAL_IP_ADDR      []byte = nil
	NETWORK_MASK       []byte = nil
	GATEWAY_IP_ADDR    []byte = nil
)

var (
	ARP_ENGINE  *ArpEngine  = nil
	ICMP_ENGINE *IcmpEngine = nil
	UDP_ENGINE  *UdpEngine  = nil
	IPV4_ENGINE *Ipv4Engine = nil
	TCP_ENGINE  *TcpEngine  = nil
)

func InitEngine(macAddr string, ipAddr string, networkMask string, gatewayIpAddr string) error {
	// local mac
	macAddrSplit := strings.Split(macAddr, ":")
	LOCAL_MAC_ADDR = make([]byte, 6)
	for i := 0; i < 6; i++ {
		split, err := strconv.ParseUint(macAddrSplit[i], 16, 8)
		if err != nil {
			return err
		}
		LOCAL_MAC_ADDR[i] = uint8(split)
	}
	// local ip
	ipAddrSplit := strings.Split(ipAddr, ".")
	LOCAL_IP_ADDR = make([]byte, 4)
	for i := 0; i < 4; i++ {
		split, err := strconv.Atoi(ipAddrSplit[i])
		if err != nil {
			return err
		}
		LOCAL_IP_ADDR[i] = uint8(split)
	}
	// network mask
	networkMaskSplit := strings.Split(networkMask, ".")
	NETWORK_MASK = make([]byte, 4)
	for i := 0; i < 4; i++ {
		split, err := strconv.Atoi(networkMaskSplit[i])
		if err != nil {
			return err
		}
		NETWORK_MASK[i] = uint8(split)
	}
	// gateway ip
	gatewayIpAddrSplit := strings.Split(gatewayIpAddr, ".")
	GATEWAY_IP_ADDR = make([]byte, 4)
	for i := 0; i < 4; i++ {
		split, err := strconv.Atoi(gatewayIpAddrSplit[i])
		if err != nil {
			return err
		}
		GATEWAY_IP_ADDR[i] = uint8(split)
	}
	protocol.SetRandIpHeaderId()
	ARP_ENGINE = NewArpEngine()
	ICMP_ENGINE = NewIcmpEngine()
	UDP_ENGINE = NewUdpEngine()
	IPV4_ENGINE = NewIpv4Engine()
	TCP_ENGINE = NewTcpEngine()
	return nil
}

func RunEngine(cpuCoreList []int, golangRxCpuCore int, memChanNum int, targetIpAddr string) {
	dpdk.Alloc()
	dpdk.Config(cpuCoreList, golangRxCpuCore, memChanNum, targetIpAddr)
	dpdk.Run()
	// 等待DPDK启动完成
	time.Sleep(time.Second * 10)
	dpdk.Handle()
	go PacketHandle()
	go ICMP_ENGINE.NetworkStateCheck()
}

func StopEngine() {
	dpdk.Exit()
}

func PacketHandle() {
	for {
		select {
		case ethFrm := <-dpdk.DPDK_RX_CHAN:
			if DEBUG {
				fmt.Printf("rx pkt, eth frm len: %v, eth frm data: %v\n", len(ethFrm), ethFrm)
			}
			ethPayload, ethDstMac, ethSrcMac, ethProto, err := protocol.ParseEthFrm(ethFrm)
			if err != nil {
				fmt.Printf("parse ethernet frame error: %v\n", err)
				continue
			}
			if !bytes.Equal(ethDstMac, BROADCAST_MAC_ADDR) && !bytes.Equal(ethDstMac, LOCAL_MAC_ADDR) {
				continue
			}
			switch ethProto {
			case protocol.ETH_PROTO_ARP:
				ARP_ENGINE.Handle(ethPayload, ethSrcMac)
			case protocol.ETH_PROTO_IP:
				IPV4_ENGINE.Handle(ethPayload)
			default:
			}
		}
	}
}
