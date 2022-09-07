package main

import (
	"dpdk-go/dpdk"
	"encoding/hex"
	"fmt"
	"time"
)

func main() {
	dpdk.Alloc()
	dpdk.Run()
	// 等待DPDK启动完成
	time.Sleep(time.Second * 10)
	//dpdk.Loopback()
	dpdk.Handle()
	go func() {
		for {
			// 接收原始以太网报文
			pkt := <-dpdk.DPDK_RX_CHAN
			fmt.Printf("rx pkt, len: %v, data: %v\n", len(pkt), pkt)
		}
	}()
	go func() {
		pkt, err := hex.DecodeString("112233aabbcc")
		if err != nil {
			panic(err)
		}
		for {
			// 发送原始以太网报文
			dpdk.DPDK_TX_CHAN <- pkt
			time.Sleep(time.Second)
		}
	}()
	time.Sleep(time.Second * 30)
	dpdk.Exit()
	select {}
}
