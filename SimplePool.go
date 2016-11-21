package TPool

import "fmt"
import "errors"
import "net"
import (
	"sync"
)

type Factory func() (ConItem,error)

type SimplePool struct {
	mutex sync.Mutex
	connection chan net.Conn
	 factory Factory
}

func NewSimplePool(size,cap int,factory Factory) (Pool,error){
	fmt.Println("fuck u every day!")
	if size < 0 && cap < 0 && factory == nil && size > cap {
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
	defer s.mutex.Unlock()
	if s.connection == nil || len(s.connection) == 0 {
		return
	}
	close(s.connection)

	for con := range s.connection {
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

func (s *SimplePool) Get() (ConItem,error){
	cons := s.GetAllConnection()
	if cons == nil || len(cons) == 0 {
		return nil,ErrClosed
	}

	select {
	case conn := <-cons:
		if cons == nil {
			return  nil,ErrClosed
		}
	default:
		conn,err := s.factory()
		if err != nil {
			return nil,err
		}
		return s.MakeConn(conn),nil
	}
}
func (s *SimplePool) MakeConn(con net.Conn) ConItem{
	p := &ConItem{pool:s}
	p.Conn = con
	return p
}
