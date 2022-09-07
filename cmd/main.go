package main

import (
	"bytes"
	"dpdk-go/dpdk"
	"dpdk-go/protocol"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	dpdk.Alloc()
	dpdk.Run()
	// 等待DPDK启动完成
	time.Sleep(time.Second * 10)
	dpdk.Handle()

	// 00:0C:29:3E:3E:DF
	localMacAddr := []byte{0x00, 0x0c, 0x29, 0x3e, 0x3e, 0xdf}
	// 192.168.199.199
	localIpAddr := []byte{0xc0, 0xa8, 0xc7, 0xc7}
	protocol.SetRandIpHeaderId()

	go func() {
		for {
			// 接收原始以太网报文
			ethFrm := <-dpdk.DPDK_RX_CHAN
			fmt.Printf("rx pkt, eth frm len: %v, eth frm data: %v\n", len(ethFrm), ethFrm)
			ethPayload, ethDstMac, ethSrcMac, ethProto, err := protocol.ParseEthFrm(ethFrm)
			if err != nil {
				fmt.Printf("parse ethernet frame error: %v", err)
				continue
			}
			if !bytes.Equal(ethDstMac, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) && !bytes.Equal(ethDstMac, localMacAddr) {
				continue
			}
			switch ethProto {
			case protocol.ETH_PROTO_ARP:
				arpOption, arpSrcMac, arpSrcAddr, _, arpDstAddr, err := protocol.ParseArpPkt(ethPayload)
				if err != nil {
					fmt.Printf("parse arp packet error: %v", err)
					continue
				}
				// 对目的IP为本机的ARP询问请求进行回应
				if arpOption == protocol.ARP_REQUEST && bytes.Equal(arpDstAddr, localIpAddr) {
					arpPkt, err := protocol.BuildArpPkt(protocol.ARP_REPLY, localMacAddr, localIpAddr, arpSrcMac, arpSrcAddr)
					if err != nil {
						fmt.Printf("build arp packet error: %v", err)
						continue
					}
					ethFrm, err := protocol.BuildEthFrm(arpPkt, ethSrcMac, localMacAddr, protocol.ETH_PROTO_ARP)
					if err != nil {
						fmt.Printf("build ethernet frame error: %v", err)
						continue
					}
					fmt.Printf("tx arp pkt, eth frm len: %v, eth frm data: %v\n", len(ethFrm), ethFrm)
					// 发送原始以太网报文
					dpdk.DPDK_TX_CHAN <- ethFrm
				}
			case protocol.ETH_PROTO_IP:
				ipv4Payload, ipv4HeadProto, ipv4SrcAddr, ipv4DstAddr, err := protocol.ParseIpv4Pkt(ethPayload)
				if err != nil {
					fmt.Printf("parse ip packet error: %v", err)
					continue
				}
				if !bytes.Equal(ipv4DstAddr, localIpAddr) {
					continue
				}
				switch ipv4HeadProto {
				case protocol.IPH_PROTO_ICMP:
					icmpPayload, icmpType, icmpId, icmpSeq, err := protocol.ParseIcmpPkt(ipv4Payload)
					if err != nil {
						fmt.Printf("parse icmp packet error: %v", err)
						continue
					}
					if icmpType == protocol.ICMP_REQUEST {
						// 构造ICMP响应包
						icmpPkt, err := protocol.BuildIcmpPkt(icmpPayload, protocol.ICMP_REPLY, icmpId, icmpSeq)
						if err != nil {
							fmt.Printf("build icmp packet error: %v", err)
							continue
						}
						ipv4Pkt, err := protocol.BuildIpv4Pkt(icmpPkt, protocol.IPH_PROTO_ICMP, localIpAddr, ipv4SrcAddr)
						if err != nil {
							fmt.Printf("build ip packet error: %v", err)
							continue
						}
						ethFrm, err := protocol.BuildEthFrm(ipv4Pkt, ethSrcMac, localMacAddr, protocol.ETH_PROTO_IP)
						if err != nil {
							fmt.Printf("build ethernet frame error: %v", err)
							continue
						}
						fmt.Printf("tx icmp pkt, eth frm len: %v, eth frm data: %v\n", len(ethFrm), ethFrm)
						dpdk.DPDK_TX_CHAN <- ethFrm
					}
				default:
				}
			default:
			}
		}
	}()

	// 至此 你已经完成了一个独立于linux内核存在的 能ping通的最简单的网络协议栈

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			time.Sleep(time.Second)
			dpdk.Exit()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
