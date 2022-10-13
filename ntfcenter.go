package mouselib

import (
	"fmt"
	"sync"
)

type Center struct {
	Status int

	Pubs        []Pub
	Subs        []Sub
	Topics      map[string]chan []byte
	PubTopicRel map[string][]string

	pubCt int
	subCt int
	m     sync.Mutex
}

func (c *Center) Start() {
	c.Status = 1
}

func (c *Center) genId(t string, ct int) string {
	return fmt.Sprintf("%s-%d", t, ct)
}

func (c *Center) RegPub(topics ...string) *Pub {
	c.m.Lock()
	defer c.m.Unlock()

	pid := c.genId("pub", c.pubCt)
	c.pubCt++

	chs := make([]chan<- []byte, len(topics))
	for i, t := range topics {
		tc := make(chan []byte, 10)
		chs[i] = tc
		c.Topics[t] = tc
		c.PubTopicRel[pid] = topics
	}
	p := Pub{
		Id:    pid,
		Chans: chs,
	}
	c.Pubs = append(c.Pubs, p)
	return &p
}

func (c *Center) UnregPub(pubId string) {
	f := -1
	for i, pub := range c.Pubs {
		if pub.Id == pubId {
			pub.Stop()
			f = i
		}
	}
	if f != -1 {
		c.m.Lock()
		defer c.m.Unlock()
		m := len(c.Pubs)
		c.Pubs[f] = c.Pubs[m-1]
		c.Pubs = c.Pubs[:m-1]

		for _, t := range c.PubTopicRel[pubId] {
			delete(c.Topics, t)
		}
	}
}

func (c *Center) RegSub(topics ...string) *Sub {
	subId := c.genId("sub", c.subCt)
	c.pubCt++

	chs := make(map[string]<-chan []byte)
	for _, t := range topics {
		chs[t] = c.Topics[t]
	}
	s := Sub{
		Id:    subId,
		Chans: chs,
	}
	c.Subs = append(c.Subs, s)
	return &s
}

func (c *Center) UnregSub(subId string) {
	f := -1
	for i, sub := range c.Subs {
		if sub.Id == subId {
			f = i
		}
	}
	if f != -1 {
		m := len(c.Subs)
		c.Subs[f] = c.Subs[m-1]
		c.Subs = c.Subs[:m-1]
	}
}

func NewCenter() *Center {
	return &Center{
		Status:      0,
		Pubs:        make([]Pub, 0),
		Subs:        make([]Sub, 0),
		Topics:      make(map[string]chan []byte),
		PubTopicRel: make(map[string][]string),
		pubCt:       0,
		subCt:       0,
		m:           sync.Mutex{},
	}
}

type Pub struct {
	Id    string
	Chans []chan<- []byte
}

func (p *Pub) Send(msg []byte) {
	for _, ch := range p.Chans {
		ch <- msg
	}
}

func (p *Pub) Stop() {
	for _, ch := range p.Chans {
		close(ch)
	}
}

type Sub struct {
	Id    string
	Chans map[string]<-chan []byte
}

func (s *Sub) Rcv() map[string]<-chan []byte {
	return s.Chans
}
