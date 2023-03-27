package utils

/*
提供一个工具，该工具可以按seq顺序返回数据
*/

import (
	"sync"
	"sync/atomic"
)

type SeqMachine struct {
	nextSeqID uint64
	ackSeqID  uint64
	msg       sync.Map
	lock      sync.Mutex
	popChan   chan interface{}
}

func NewSeqMachine(initSeqID uint64) *SeqMachine {
	return &SeqMachine{
		nextSeqID: initSeqID,
		ackSeqID:  initSeqID + 1,
		msg:       sync.Map{},
		lock:      sync.Mutex{},
		popChan:   make(chan interface{}, 64),
	}
}

// 顺序发号
func (s *SeqMachine) Push() uint64 {
	return atomic.AddUint64(&s.nextSeqID, 1)
}

// 更新为已确认
func (s *SeqMachine) Update(seqID uint64, msg interface{}) {
	s.msg.Store(seqID, msg)

	s.lock.Lock()
	defer s.lock.Unlock()

	for {
		msg, ok := s.msg.Load(s.ackSeqID)
		if !ok {
			return
		}

		s.popChan <- msg

		s.msg.Delete(s.ackSeqID)
		s.ackSeqID++
	}
}

// 取出数据
func (s *SeqMachine) Pop() interface{} {
	return <-s.popChan
}

// 返回pop管道
func (s *SeqMachine) PopChan() <-chan interface{} {
	return s.popChan
}
