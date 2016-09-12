package cache

// An AtomicInt is an int64 to be accessed atomically.
import (
	"strconv"
	"sync/atomic"
)

type AtomicInt int64

// Add atomically adds n to i.
func (i *AtomicInt) Add(n int64) {
	atomic.AddInt64((*int64)(i), n)
}

// Get atomically gets the value of i.
func (i *AtomicInt) Get() int64 {
	return atomic.LoadInt64((*int64)(i))
}

func (i *AtomicInt) String() string {
	return strconv.FormatInt(i.Get(), 10)
}

// CacheStats are returned by stats accessors on Group.
type CacheStats struct {
	Bytes     AtomicInt
	Items     AtomicInt
	LGets     AtomicInt
	LHits     AtomicInt
	RGets     AtomicInt
	RHits     AtomicInt
	Evictions AtomicInt
}

// Stats are per-group statistics.
type Stats struct {
	Gets          AtomicInt // any Get request, including from peers
	CacheHits     AtomicInt // either cache was good
	Loads         AtomicInt // (gets - cacheHits)
	LoadsDeduped  AtomicInt // after singleflight
	LocalLoads    AtomicInt // total good local loads
	LocalLoadErrs AtomicInt // total bad local loads
}
