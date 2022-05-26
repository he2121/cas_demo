package main

import (
	"math/rand"
	"time"
)

var memoryTicket = map[string]*Ticket{}

type Ticket struct {
	Ticket  string
	Type    string
	Service string
	Created time.Time
}

func NewServiceTicket(service string) *Ticket {
	ticket := createNewTicket("ST")
	ticket.Service = service
	memoryTicket[ticket.Ticket] = &ticket
	return &ticket
}

func NewTGT() *Ticket {
	ticket := createNewTicket("TGC")
	memoryTicket[ticket.Ticket] = &ticket
	return &ticket
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

func createNewTicket(ticketType string) Ticket {
	c := 100
	b := make([]byte, c)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}

	return Ticket{
		Ticket:  ticketType + "-" + string(b),
		Type:    ticketType,
		Created: time.Now(),
	}
}
