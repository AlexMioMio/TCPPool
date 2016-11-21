package TPool

import "testing"
import "log"
import "net"
import "sync"
import "math/rand"
import (
	"time"
	"fmt"
)

var (
	size = 5
	cap = 30
	network = "tcp"
	address = "127.0.0.1:7777"
	factory = func () (net.Conn,error) {return net.Dial(network,address)}
)
func ReadConnection(con net.Conn){
	buffer := make([]byte,256)
	con.Read(buffer)
}

func SimpleServer() {
	listener,err := net.Listen(network,address)
	if err != nil {
		log.Fatal("build server error")
	}
	defer listener.Close()
	for {
		conn,err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go ReadConnection(conn)
	}
}

func init() {
	go SimpleServer()
	time.Sleep(time.Second * 5)
	rand.Seed(time.Now().UTC().UnixNano())
}

func Test_NewSimplePool(t *testing.T){
	_,err := NewSimplePool(size,cap,factory)

	if err != nil {
		t.Errorf("create connection:",err)
	}
}

func Test_Get(t *testing.T) {
	pool,err := NewSimplePool(size,cap,factory)
	defer  pool.Close()
	if err != nil || pool == nil {
		t.Errorf("create connetion pool:",err)
	}

	con,err := pool.Get()
	if err != nil  {
		t.Errorf("Get connection failed:",err)
	}
	_,ok := con.(*ConItem)
	if !ok {
		t.Errorf("fuck err")
	}
}

func Test_GetSync(t *testing.T){
	pool,err := NewSimplePool(size,cap,factory)
	defer pool.Close()
	if err != nil {
		t.Errorf("create connection pool failed")
	}

	if pool.Len() != size -1 {
		t.Errorf("expect len:%d,but get %d",size,pool.Len())
	}

	var wg sync.WaitGroup
	for i := 0 ; i < size ; i++ {
		wg.Add(1)
		go func(){
			defer wg.Done()
			_,err := pool.Get()
			if err != nil {
				t.Errorf("error:",err)
			}
		}()
	}
	wg.Wait()

	if pool.Len() != 0 {
		t.Errorf("Expect len=0,but get %d",pool.Len())
	}
	_ ,err = pool.Get()
	if err != nil {
		t.Errorf("Get error: %s",err)
	}
}

func Test_Put(t *testing.T) {
	pool,err := NewSimplePool(size,cap,factory)
	defer pool.Close()
	if err != nil {
		t.Errorf("create connection pool,but meet:",err)
	}

	cons := make([]net.Conn,cap)
	for i := 0 ; i < cap; i++{
		conn,_ := pool.Get()
		cons[i] = conn
	}
	for _, conn := range cons {
		conn.Close()
	}
	if pool.Len() != cap {
		fmt.Println(pool.Len())
		t.Errorf("len is not correct")
	}

	conn, _ := pool.Get()
	pool.Close()
	conn.Close()
	if pool.Len() != 0 {
		t.Errorf("expect len is 0")
	}
}

func Test_PutConn(t *testing.T) {
	pool , _ := NewSimplePool(size,cap,factory)
	defer pool.Close()

	conn,_ := pool.Get()
	conn.Close()
	pSize := pool.Len()
	conn, _ = pool.Get()
	conn.Close()
	if pool.Len() != pSize {
		t.Errorf("we expect len is %d",pSize)
	}

	conn,_ = pool.Get()
	if pc,ok := conn.(*ConItem) ; !ok {
		t.Errorf("transform error")
	} else {
		pc.MarkUnused()
	}
	conn.Close()
	if pool.Len()  != size -1 {
		t.Errorf("expect size %d,but get %d",size-1,pool.Len())
	}
}

func Test_Cap(t *testing.T){
	pool,_ := NewSimplePool(size,cap,factory)
	defer  pool.Close()

	if pool.Len() != cap {
		t.Errorf("")
	}
}

func Test_Close(t *testing.T){
	pool, _ := NewSimplePool(size,cap,factory)
	defer pool.Close()

	pool.Close()
	c := pool.(*SimplePool)

	if c.connection != nil {
		t.Errorf("channel should be closed")
	}
	if c.factory != nil {
		t.Errorf("factory should be closed")
	}
	_,err := pool.Get()
	if err == nil {
		t.Errorf("that's insane,should not get a connection")
	}
	if pool.Len() != 0 {
		t.Errorf("Close error,should not exit channel")
	}
}

func Test_func(t *testing.T){
	pool,_ := NewSimplePool(size,cap,factory)
	defer pool.Close()

	pipe := make(chan net.Conn,0)

	go func() {
		pool.Close()
	}()

	for i := 0 ; i < cap ; i++ {
		go func(){
			conn, _ := pool.Get()

			pipe <- conn
		}()

		go func() {
			conn := <-pipe
			if conn == nil {
				return
			}
			conn.Close()
		}()
	}
}

func Test_func2(t *testing.T) {
	pool,_ := NewSimplePool(0,30,factory)
	defer pool.Close()

	conn,_ := pool.Get()

	msg := "hello"
	_,err := conn.Write([]byte(msg))
	if err != nil {
		t.Error(err)
	}
}

func Test_func3(t *testing.T){
	pool,_ := NewSimplePool(0,30,factory)
	defer pool.Close()

	var wg sync.WaitGroup

	go func(){
		for i := 0 ; i <10 ; i++ {
			wg.Add(1)
			go func(){
				conn,_ := pool.Get()
				time.Sleep(time.Second * 1)
				conn.Close()
				wg.Done()
			}()
		}
	}()

	for i := 0 ; i < 10 ; i++ {
		wg.Add(1)
		go func(){
			conn,_ := pool.Get()
			time.Sleep(time.Second*1)
			conn.Close()
			wg.Done()
		}()
	}
	wg.Wait()
}



