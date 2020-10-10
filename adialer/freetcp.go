package adialer

import (
	"context"
	"encoding/json"
	"net"
	"sync"

	"github.com/rikaaa0928/awsl/aconn"
	"github.com/rikaaa0928/awsl/consts"
	"github.com/rikaaa0928/awsl/utils"
)

var FreeTCP = func(ctx context.Context, addr net.Addr) (context.Context, aconn.AConn, error) {
	c, err := net.Dial("tcp", addr.String())
	ac := aconn.NewAConn(c)
	ac.SetEndAddr(addr)
	return ctx, ac, err
}

var FreeUDP = func(ctx context.Context, src, dst string) (context.Context, aconn.AConn, error) {
	uDst, err := net.ResolveUDPAddr("udp", dst)
	if err != nil {
		return ctx, nil, err
	}
	//uSrc, err := net.ResolveUDPAddr("udp", src)
	//if err != nil {
	//	return ctx, nil, err
	//}
	c, err := net.DialUDP("udp", nil, uDst)
	if err != nil {
		return ctx, nil, err
	}
	ac := aconn.NewAConn(&udpConnWrapper{UDPConn: c, toAddr: uDst, src: src, dst: dst})
	ac.SetEndAddr(uDst)
	return ctx, ac, err
}

type udpConnWrapper struct {
	sync.Mutex
	*net.UDPConn
	toAddr *net.UDPAddr
	src    string
	dst    string
}

func (c *udpConnWrapper) reDial() error {
	c2, err := net.DialUDP("udp", nil, c.toAddr)
	if err != nil {
		return err
	}
	c.UDPConn = c2
	return nil
}

func (c *udpConnWrapper) Read(b []byte) (n int, err error) {
	buf := utils.GetMem(65536)
	defer utils.PutMem(buf)
	var dstAddr *net.UDPAddr
	n, dstAddr, err = c.UDPConn.ReadFromUDP(buf)
	i := 0
	for err != nil && i < 3 {
		err = c.reDial()
		if err != nil {
			continue
		}
		n, dstAddr, err = c.UDPConn.ReadFromUDP(buf)
		if err != nil {
			break
		}
		i++
	}
	if err != nil {
		return
	}
	udp := consts.UDPMSG{
		DstStr: dstAddr.String(),
		SrcStr: c.src,
		Data:   buf[:n],
	}
	str, err := json.Marshal(udp)
	if err != nil {
		return
	}
	n = len(str)
	copy(b, str)
	return
}

func (c *udpConnWrapper) Write(b []byte) (n int, err error) {
	var udpMsg consts.UDPMSG
	err = json.Unmarshal(b, &udpMsg)
	if err != nil {
		return -1, err
	}
	n, err = c.UDPConn.Write(udpMsg.Data)
	i := 0
	for err != nil && i < 3 {
		err = c.reDial()
		if err != nil {
			continue
		}
		n, err = c.UDPConn.Write(udpMsg.Data)
		if err != nil {
			break
		}
		i++
	}
	return
}