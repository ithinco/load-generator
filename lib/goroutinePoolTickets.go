package lib

import (
	"errors"
	"fmt"
)

type GoroutinePoolTickets interface {
	Take()
	PutBack()
	Active() bool
	Total() uint64
	RemainingTickets() uint64
}

type goroutinePoolTickets struct {
	total uint64
	ticketChan chan struct{}
	active bool
}

func (receiver *goroutinePoolTickets) init(total uint64) bool  {
	if receiver.active {
		return false
	}
	if total == 0 {
		return false
	}

	ch := make(chan struct{}, total)
	for i := 0; i < int(total); i++ {
		ch <- struct{}{}
	}

	receiver.ticketChan = ch
	receiver.total = total
	receiver.active = true

	return true
}

func (receiver *goroutinePoolTickets) Take()  {
	<-receiver.ticketChan
}

func (receiver *goroutinePoolTickets) PutBack()  {
	receiver.ticketChan<- struct{}{}
}

func (receiver *goroutinePoolTickets) Active() bool  {
	return receiver.active
}

func (receiver *goroutinePoolTickets) RemainingTickets() uint64  {
	return uint64(len(receiver.ticketChan))
}

func (receiver *goroutinePoolTickets) Total() uint64  {
	return receiver.total
}

func NewGoroutinePoolTickets(total uint64) (GoroutinePoolTickets, error) {
	gpt := goroutinePoolTickets{}
	if !gpt.init(total) {
		errMsg := fmt.Sprintf("Can't initialize goroutinePoolTickets with total=%d",total)
		return nil, errors.New(errMsg)
	}
	return &gpt, nil
}