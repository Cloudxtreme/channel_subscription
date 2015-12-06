// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by Apache 2.0
// license that can be found in the LICENSE file.

// Package channel_subscription provides a way to send from 1 to many channels.
package channel_subscription

import (
	"errors"
	"reflect"
	"sync"
)

//ErrNotFound is an error.
var ErrNotFound = errors.New("subscription not found")

// ChannelSubscription map stores all channels to send.
type ChannelSubscription struct {
	channels map[reflect.Value]struct{}
	lck      sync.Mutex
}

func NewChannelSubscription() *ChannelSubscription {
	return &ChannelSubscription{
		channels: make(map[reflect.Value]struct{}),
	}
}

// Subscribe subscribes an remote channel that will send data to ch channel.
// ch channel must be of one type only.
func (c *ChannelSubscription) Subscribe(ch interface{}) {
	c.lck.Lock()
	defer c.lck.Unlock()
	val := reflect.ValueOf(ch)
	if val.Kind() != reflect.Chan {
		panic("ch type isn't channel")
	}
	c.channels[val] = struct{}{}
}

// Unsubscribe removes the channel ch from the pool, nothing will be
// send to it.
func (c *ChannelSubscription) Unsubscribe(ch interface{}) error {
	c.lck.Lock()
	defer c.lck.Unlock()
	val := reflect.ValueOf(ch)
	_, found := c.channels[val]
	if !found {
		return ErrNotFound
	}
	delete(c.channels, val)
	return nil
}

// TrySend sends data to the channels subscribeds. i type must be
// compatible with ch type. The semantics is the same of the
// TrySend method in the reflect.Value type.
func (c *ChannelSubscription) TrySend(i interface{}) (sent bool) {
	c.lck.Lock()
	defer c.lck.Unlock()
	sent = true
	val := reflect.ValueOf(i)
	for ch := range c.channels {
		sent = sent && ch.TrySend(val)
	}
	return
}

// Send sends data to the channels subscribeds. i type must be
// compatible with ch type. The semantics is the same of the
// Send method in the reflect.Value type.
func (c *ChannelSubscription) Send(i interface{}) {
	c.lck.Lock()
	defer c.lck.Unlock()
	val := reflect.ValueOf(i)
	var wg sync.WaitGroup
	wg.Add(len(c.channels))
	for ch := range c.channels {
		go func(ch reflect.Value) {
			defer wg.Done()
			ch.Send(val)
		}(ch)
	}
	wg.Wait()
}
