package TPool

import "fmt"
import "errors"
import "net"
import (
	"sync"
)

type Factory func() (net.Conn,error)

type SimplePool struct {
	mutex sync.Mutex
	connection chan net.Conn
	factory Factory
}

func NewSimplePool(size,cap int,factory Factory) (Pool,error){
	fmt.Println("make simplePool")
	if size < 0 || cap <= 0 || size > cap {
		return nil,errors.New("invalid params")
	}

	c := &SimplePool{
		connection: make(chan  net.Conn,cap),
		factory:factory,
	}

	for i := 0 ; i < size ; i++ {
		conn,err := factory()
		if err != nil {
			c.Close()
			return nil,errors.New("build connection,meet error")
		}
		c.connection <- conn
	}
	return c,nil
}

func (s *SimplePool) GetAllConnection()chan net.Conn{
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.connection
}
func (s *SimplePool) Len() int{
	s.mutex.Lock()
	defer s.mutex.Unlock()
	lens := len(s.connection)
	return lens
}

func (s *SimplePool) Close() {
	s.mutex.Lock()
	connection := s.connection
	s.connection = nil
	s.factory = nil
	s.mutex.Unlock()
	if connection == nil {
		return
	}
	close(connection)

	for con := range connection {
		con.Close()
	}
}

func (s *SimplePool) Put(con net.Conn) error {
	if con == nil {
		return errors.New("connection is nil")
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.connection == nil {
		con.Close()
		return errors.New("connection chan is close")
	}

	select {
	case s.connection <- con:
		return nil
	default:
		con.Close()
		return errors.New("connection chan is close")
	}
}

func (s *SimplePool) Get() (net.Conn,error){
	cons := s.GetAllConnection()
	if cons == nil {
		return nil,ErrClosed
	}

	select {
	case conn := <-cons:
		if conn == nil {
			return  nil,ErrClosed
		}
		return s.MakeConn(conn),nil
	default:
		conn,err := s.factory()
		if err != nil {
			return nil,err
		}
		return s.MakeConn(conn),nil
	}
}
func (s *SimplePool) MakeConn(con net.Conn) net.Conn{
	p := &ConItem{pool:s}
	p.Conn = con
	return p
}
