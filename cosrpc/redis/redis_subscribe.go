package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/hwcer/cosgo/logger"
)

type subscribe struct {
	pubsub  *redis.PubSub
	closeCh chan struct{}
}

func newSubscribe(client *redis.Client, regex string) (*subscribe, error) {
	ch := client.PSubscribe(context.Background(), regex)
	return &subscribe{
		pubsub:  ch,
		closeCh: make(chan struct{}),
	}, nil
}

func (s *subscribe) Close() error {
	close(s.closeCh)
	if err := s.pubsub.Close(); err != nil {
		logger.Debug("redis discovery subscribe Close error:%v", err)
	}
	return nil
}

func (s *subscribe) Receive(stopCh <-chan struct{}) chan *redis.Message {
	msgCh := make(chan *redis.Message)
	go s.receiveLoop(msgCh, stopCh)
	return msgCh
}

func (s *subscribe) receiveLoop(msgCh chan *redis.Message, stopCh <-chan struct{}) {
	defer close(msgCh)
	defer s.pubsub.Close()
	for {
		select {
		case <-s.closeCh:
			return
		case <-stopCh:
			return
		default:
			msg, err := s.pubsub.ReceiveMessage(context.Background())
			if err != nil {
				return
			}
			if msg != nil {
				msgCh <- msg
			}
		}
	}
}
