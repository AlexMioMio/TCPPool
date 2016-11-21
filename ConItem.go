package TPool

import (
	"net"
	"fmt"
)

type ConItem struct {
	unUsed bool
	net.Conn
	pool *SimplePool
}

func (c *ConItem) Close() error {
	if c.unUsed {
		fmt.Println("fuck")
		if c.Conn != nil {
			return c.Conn.Close()
		}
		return nil
	}
	err := c.pool.Put(c.Conn)
	return err
}

func (c *ConItem) MarkUnused() {
	c.unUsed = true
}



