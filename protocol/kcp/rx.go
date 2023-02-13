package kcp

func (s *UDPSession) readLoop() {
	buf := make([]byte, mtuLimit)
	var src string
	for {
		if n, addr, err := s.conn.ReadFrom(buf); err == nil {
			udpPayload := buf[:n]

			// make sure the packet is from the same source
			if src == "" { // set source address
				src = addr.String()
			} else if addr.String() != src {
				// atomic.AddUint64(&DefaultSnmp.InErrs, 1)
				// continue
				s.remote = addr
				src = addr.String()
			}

			if n == 20 {
				connType, _, conv, err := ParseEnet(udpPayload)
				if err != nil {
					continue
				}
				if conv != s.GetConv() {
					continue
				}
				if connType == ConnEnetFin {
					s.Close()
					continue
				}
			}

			s.packetInput(udpPayload)
		} else {
			s.notifyReadError(err)
			return
		}
	}
}

func (l *Listener) monitor() {
	buf := make([]byte, mtuLimit)
	for {
		if n, from, err := l.conn.ReadFrom(buf); err == nil {
			udpPayload := buf[:n]
			var convId uint64 = 0
			if n == 20 {
				connType, enetType, conv, err := ParseEnet(udpPayload)
				if err != nil {
					continue
				}
				convId = conv
				switch connType {
				case ConnEnetSyn:
					// 客户端前置握手获取conv
					l.EnetNotify <- &Enet{
						Addr:     from.String(),
						ConvId:   convId,
						ConnType: ConnEnetSyn,
						EnetType: enetType,
					}
				case ConnEnetEst:
					// 连接建立
					l.EnetNotify <- &Enet{
						Addr:     from.String(),
						ConvId:   convId,
						ConnType: ConnEnetEst,
						EnetType: enetType,
					}
				case ConnEnetFin:
					// 连接断开
					l.EnetNotify <- &Enet{
						Addr:     from.String(),
						ConvId:   convId,
						ConnType: ConnEnetFin,
						EnetType: enetType,
					}
				default:
					continue
				}
			} else {
				// 正常KCP包
				convId += uint64(udpPayload[0]) << 0
				convId += uint64(udpPayload[1]) << 8
				convId += uint64(udpPayload[2]) << 16
				convId += uint64(udpPayload[3]) << 24
				convId += uint64(udpPayload[4]) << 32
				convId += uint64(udpPayload[5]) << 40
				convId += uint64(udpPayload[6]) << 48
				convId += uint64(udpPayload[7]) << 56
			}
			l.sessionLock.RLock()
			conn, exist := l.sessions[convId]
			l.sessionLock.RUnlock()
			if exist {
				if conn.remote.String() != from.String() {
					conn.remote = from
					// 连接地址改变
					l.EnetNotify <- &Enet{
						Addr:     conn.remote.String(),
						ConvId:   convId,
						ConnType: ConnEnetAddrChange,
					}
				}
			}
			l.packetInput(udpPayload, from, convId)
		} else {
			l.notifyReadError(err)
			return
		}
	}
}
