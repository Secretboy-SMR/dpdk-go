package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/FlourishingWorld/dpdk-go/dpdk"
	"github.com/FlourishingWorld/dpdk-go/engine"
	"github.com/FlourishingWorld/dpdk-go/protocol/kcp"
)

func main() {
	// 一个独立于linux内核存在的协议栈
	err := engine.InitEngine("00:0C:29:3E:3E:DF", "192.168.199.199", "255.255.255.0", "192.168.199.1")
	if err != nil {
		panic(err)
	}
	engine.RunEngine([]int{0, 1, 2, 3}, 1, "0.0.0.0")
	// 等待基础驱动模块启动完成
	time.Sleep(time.Second * 30)

	KcpServer()
	KcpClient()

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

func KcpServer() {
	go func() {
		listener, err := kcp.ListenWithOptions("0.0.0.0:22222")
		if err != nil {
			panic(err)
		}
		go func() {
			for {
				enetNotify := <-listener.EnetNotify
				if enetNotify.ConnType == kcp.ConnEnetSyn {
					listener.SendEnetNotifyToPeer(&kcp.Enet{
						Addr:     enetNotify.Addr,
						ConvId:   1234567890123456789,
						ConnType: kcp.ConnEnetEst,
						EnetType: enetNotify.EnetType,
					})
				}
			}
		}()
		conn, err := listener.AcceptKCP()
		if err != nil {
			panic(err)
		}
		for {
			buf := make([]byte, 1472)
			size, err := conn.Read(buf)
			if err != nil {
				panic(err)
			}
			buf = buf[:size]
			fmt.Printf("recv kcp data: %v\n", buf)
			_, err = conn.Write([]byte{0x01, 0x23, 0xcd, 0xef})
			if err != nil {
				panic(err)
			}
		}
	}()
}

func KcpClient() {
	go func() {
		conn, err := kcp.DialWithOptions("192.168.199.199:22222", "0.0.0.0:30000")
		if err != nil {
			panic(err)
		}
		for {
			time.Sleep(time.Second)
			_, err = conn.Write([]byte{0x45, 0x67, 0x89, 0xab})
			if err != nil {
				panic(err)
			}
			buf := make([]byte, 1472)
			size, err := conn.Read(buf)
			if err != nil {
				panic(err)
			}
			buf = buf[:size]
			fmt.Printf("recv kcp data: %v\n", buf)
		}
	}()
}
