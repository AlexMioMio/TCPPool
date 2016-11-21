package TPool

import (
	"net"
)

type ConItem struct {
	isUsed bool
	net.Conn
	pool *SimplePool
}

func (c *ConItem) Close() error {
	if c.isUsed {
		if c.Conn != nil {
			return c.Conn.Close()
		}
		return nil
	}
	err := c.pool.Put(c.Conn)
	return err
}

func (c *ConItem) MarkUnused() {
	c.isUsed = true
}


