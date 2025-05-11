package subpub

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const subject = "s"
const msg = "msg"

func TestSubscribe(t *testing.T) {
	subpub := NewEventChannel()

	// Неверные параметры - подписка невозможна
	subscriber, err := subpub.Subscribe("", func(interface{}) {})
	assert.Error(t, err, errors.New("Subject cant be empty"))
	assert.Empty(t, subscriber)

	subscriber, err = subpub.Subscribe(subject, nil)
	assert.Error(t, err, errors.New("MessageHandler cant be nil"))
	assert.Empty(t, subscriber)
	assert.Empty(t, len(subpub.sub[subject]))

	// Успешная  подписка
	subscriber, err = subpub.Subscribe(subject, func(interface{}) {})
	assert.Empty(t, err)
	assert.Equal(t, len(subpub.sub[subject]), 1)
}

func TestRemoveSubscriber(t *testing.T) {
	subpub := NewEventChannel()
	channel1 := make(chan interface{}, 1)
	channel2 := make(chan interface{}, 1)
	subpub.sub[subject] = []chan interface{}{channel1, channel2}

	subpub.removeSubscriber(subject, channel2)
	assert.Equal(t, subpub.sub[subject][0], channel1)
	subpub.removeSubscriber(subject, channel1)
	assert.Empty(t, subpub.sub[subject])
	subpub.removeSubscriber(subject, channel1) // Не возникает паника
}

func TestPublish(t *testing.T) {
	subpub := NewEventChannel()
	assert.Empty(t, subpub.Publish(subject, "")) // Нет подписок

	channel1 := make(chan interface{}, 1)
	channel2 := make(chan interface{}, 1)
	subpub.sub[subject] = []chan interface{}{channel1, channel2}
	subpub.Publish(subject, msg)
	msg1 := <-channel1
	msg2 := <-channel2
	assert.Equal(t, msg1, msg)
	assert.Equal(t, msg2, msg)
}

func TestClose(t *testing.T) {
	subpub := NewEventChannel()
	channel1 := make(chan interface{}, 1)
	channel2 := make(chan interface{}, 1)
	subpub.sub[subject] = []chan interface{}{channel1, channel2}
	// Контекст отменился раньше чем удалились подписки
	ctx, _ := context.WithTimeout(context.Background(), time.Nanosecond)
	subpub.Close(ctx)
	assert.Equal(t, subpub.sub[subject][0], channel1)
	assert.Equal(t, subpub.sub[subject][1], channel2)

	// Контекст отменился позже, поэтому подписки успели удалиться
	ctx, _ = context.WithTimeout(context.Background(), 3*time.Second)
	subpub.Close(ctx)
	assert.Empty(t, subpub.sub[subject])
}
