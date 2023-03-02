package main

import (
	"context"
	"sync"
)

type store struct {
	sync.Mutex
	subs map[*subscriber]struct{}
}

func newStore() *store {
	return &store{
		subs: map[*subscriber]struct{}{},
	}
}

func (s *store) publish(ctx context.Context, msg *message) error {
	s.Lock()
	for sub := range s.subs {
		sub.publish(ctx, msg)
	}
	s.Unlock()
	return nil
}

func (s *store) subscribe(ctx context.Context, sub *subscriber) error {
	s.Lock()
	s.subs[sub] = struct{}{}
	s.Unlock()

	go func() {
		select {
		case <-sub.quit:
		case <-ctx.Done():
			s.Lock()
			delete(s.subs, sub)
			s.Unlock()
		}
	}()

	go sub.run(ctx)
	return nil
}

func (s *store) unsubscribe(ctx context.Context, sub *subscriber) error {
	s.Lock()
	delete(s.subs, sub)
	s.Unlock()
	close(sub.quit)
	return nil
}

func (s *store) subscribes() int {
	s.Lock()
	n := len(s.subs)
	s.Unlock()
	return n
}
