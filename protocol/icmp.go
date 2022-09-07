package protocol

import (
	"errors"
)

/*
					ICMP报文
0						2						4(字节)
+-----------------------------------------------+
|	类型		|	代码		|		校验和			|
+-----------------------------------------------+
|			标识			|		序号				|
+-----------------------------------------------+
|						数据						|
+-----------------------------------------------+
*/

var ICMP_DEFAULT_PAYLOAD = []byte{
	0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0x6a, 0x6b, 0x6c, 0x6d, 0x6e, 0x6f, 0x70,
	0x71, 0x72, 0x73, 0x74, 0x75, 0x76, 0x77, 0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69,
}

const (
	ICMP_REQUEST uint8 = 0x08
	ICMP_REPLY   uint8 = 0x00
	ICMP_UNKNOWN uint8 = 0xff
)

func ParseIcmpPkt(pkt []byte) (payload []byte, icmpType uint8, icmpId []byte, icmpSeq []byte, err error) {
	if len(pkt) < 8 || len(pkt) > 1480 {
		return nil, ICMP_UNKNOWN, nil, nil, errors.New("icmp packet len must >= 8 and <= 1480 bytes")
	}
	// 类型
	icmpType = pkt[0]
	// 标识
	icmpId = pkt[4:6]
	// 序号
	icmpSeq = pkt[6:8]
	// 数据
	payload = pkt[8:]
	return payload, icmpType, icmpId, icmpSeq, nil
}

func BuildIcmpPkt(payload []byte, icmpType uint8, icmpId []byte, icmpSeq []byte) (pkt []byte, err error) {
	if len(payload) > 1472 {
		return nil, errors.New("payload len must <= 1472")
	}
	pkt = make([]byte, 0)
	// 类型
	pkt = append(pkt, icmpType)
	// 代码
	pkt = append(pkt, 0x00)
	// 校验和(暂时用0代替)
	pkt = append(pkt, 0x00, 0x00)
	// 标识
	pkt = append(pkt, icmpId...)
	// 序号
	pkt = append(pkt, icmpSeq...)
	// 数据
	pkt = append(pkt, payload...)
	// 校验和
	sum := getCheckSum(pkt)
	pkt[2] = sum[1]
	pkt[3] = sum[0]
	return pkt, nil
}
