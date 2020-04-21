package clients

import (
	"encoding/json"
	"errors"
	"net"
	"time"

	"github.com/Evi1/awsl/model"
	"github.com/Evi1/awsl/tools/crypt"
	"github.com/Evi1/awsl/tools/dialer"
)

// NewTCP NewTCP
func NewTCP(host, port, auth string, backup []string) *TCP {
	m := make(map[string][]string)
	hp := net.JoinHostPort(host, port)
	m[hp] = []string{hp}
	if backup != nil {
		m[hp] = append(m[hp], backup...)
	}
	d := &dialer.MultiAddr{Hosts: m, HostInUse: make(map[string]uint)}
	return &TCP{ServerHost: host, ServerPort: port, Auth: auth, Dialer: d}
}

// TCP tcp
type TCP struct {
	ServerHost string
	ServerPort string
	Auth       string
	Dialer     *dialer.MultiAddr
}

// Dial dial
func (c *TCP) Dial(addr model.ANetAddr) (net.Conn, error) {
	conn, err := c.Dialer.Dial("tcp", net.JoinHostPort(c.ServerHost, c.ServerPort))
	//conn, err := net.Dial("tcp", net.JoinHostPort(c.ServerHost, c.ServerPort))
	if err != nil {
		return conn, err
	}
	//return conn, nil
	return cryptConn{Conn: conn, Addr: addr, cry: crypt.Simple(1)}, nil
}

// Verify verify
func (c *TCP) Verify(conn net.Conn) error {
	ws, ok := conn.(cryptConn)
	if !ok {
		return errors.New("wrong type")
	}
	auth := model.AddrWithAuth{ANetAddr: ws.Addr, Auth: c.Auth}
	addrBytes, err := json.Marshal(auth)
	if err != nil {
		//log.Println("json marshal : " + err.Error())
		return err
	}
	_, err = ws.Write(addrBytes)
	time.Sleep(10 * time.Millisecond)
	return err
}

type cryptConn struct {
	net.Conn
	Addr model.ANetAddr
	cry  crypt.Cryptor
}

func (c cryptConn) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)
	c.cry.Decrypt(b, n)
	return
}

func (c cryptConn) Write(b []byte) (n int, err error) {
	c.cry.Encrypt(b, len(b))
	n, err = c.Conn.Write(b)
	return
}
