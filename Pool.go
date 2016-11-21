package TPool

import "errors"
import "net"

var ErrClosed = errors.New("Pool is closed")

type Pool interface {
	Get() (net.Conn,error)
	Close()
	Len() int
}
