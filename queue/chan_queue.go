package queue

import (
    "sync/atomic"
)

type ChanQueue struct {
    syncChan chan interface{}
    size int32
}

type ChanStop struct {}

func NewChanQueue(maxSize int32) *ChanQueue {
    return &ChanQueue{syncChan: make(chan interface{}, maxSize)}
}

func (q *ChanQueue) GetChan() *chan interface{} {
    return &q.syncChan
}

func (q *ChanQueue) Push(v interface{}) {
    atomic.AddInt32(&q.size, 1)
    q.syncChan <- v
}

/*
    If sizes is not zero, it will return true, else will return false.
*/
func (q *ChanQueue) Pop() (interface{}, bool) {
    if q.Len() <= 0 {
        return nil, false
    }
    atomic.AddInt32(&q.size, -1)
    v := <-q.syncChan
    return v, true
}

func (q *ChanQueue) Len() int {
    return int(q.size)
}


/*
    Non-thread safe method.
*/
func (q *ChanQueue) Push_th(v interface{}) {
    q.syncChan <- v
}
func (q *ChanQueue) Pop_th() (interface{}, bool) {
    if len(q.syncChan) <= 0 {
        return nil, false
    }
    v := <-q.syncChan
    return v, true
}
func (q *ChanQueue) Len_th() int {
    return len(q.syncChan)
}
