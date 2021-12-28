package EPUtils

import "sync/atomic"

type Count32 int32

func (c *Count32) Inc() int32 {
	return atomic.AddInt32((*int32)(c), 1)
}

func (c *Count32) Dec() int32 {
	return atomic.AddInt32((*int32)(c), -1)
}

func (c *Count32) Get() int32 {
	return atomic.LoadInt32((*int32)(c))
}
