// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by Apache 2.0
// license that can be found in the LICENSE file.

package channel_subscription

import (
	"errors"
	"reflect"
	"sync"
)

var ErrNotFound = errors.New("subscription not found")

type ChannelSubscription map[reflect.Value]bool

func (c ChannelSubscription) Subscribe(ch interface{}) {
	val := reflect.ValueOf(ch)
	if val.Kind() != reflect.Chan {
		panic("ch type isn't channel")
	}
	c[val] = true
}

func (c ChannelSubscription) Unsubscribe(ch interface{}) error {
	val := reflect.ValueOf(ch)
	_, found := c[val]
	if !found {
		return ErrNotFound
	}
	delete(c, val)
	return nil
}

func (c ChannelSubscription) TrySend(i interface{}) (sent bool) {
	sent = true
	val := reflect.ValueOf(i)
	for ch := range c {
		sent = sent && ch.TrySend(val)
	}
	return
}

func (c ChannelSubscription) Send(i interface{}) {
	val := reflect.ValueOf(i)
	var wg sync.WaitGroup
	wg.Add(len(c))
	for ch := range c {
		go func(ch reflect.Value) {
			defer wg.Done()
			ch.Send(val)
		}(ch)
	}
	wg.Wait()
}
