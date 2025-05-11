package subpub

import (
	"context"
	"fmt"
	"sync"
)

// Структура шины событий
type EventChannel struct {
	sub map[string][]chan interface{} // Подписки
	// Для блокировки одновременного доступа к sub
	mutex sync.RWMutex
	wg    sync.WaitGroup // Отслеживание количества подписчиков
}

const sizeBufferChan = 10

func NewEventChannel() *EventChannel {
	return &EventChannel{
		sub: make(map[string][]chan interface{}),
		wg:  sync.WaitGroup{},
	}
}

// Функция для подписки на subject и вызыва для него обработчика cb
// Возврат Subscription для отмены подписки
func (e *EventChannel) Subscribe(subject string, cb MessageHandler) (
	Subscription, error) {
	if subject == "" {
		return nil, fmt.Errorf("Subject cant be empty")
	}
	if cb == nil {
		return nil, fmt.Errorf("MessageHandler cant be nil")
	}
	channel := make(chan interface{}, sizeBufferChan)
	e.mutex.Lock()
	e.sub[subject] = append(e.sub[subject], channel)
	e.wg.Add(1)
	e.mutex.Unlock()
	subscriber := NewSubscriber(e, subject, channel)
	go func() {
		defer e.wg.Done()
		for {
			select {
			// Ожидание сообщения или закрытия канала
			case msg, ok := <-channel:
				if !ok { // Канал закрыт
					return
				}
				cb(msg) // Вызов обработчика для сообщения msg
			}
		}
	}()
	return subscriber, nil
}

// Функция для удаления подписки
// Должно гарантироваться, что не закрывается повторно
func (e *EventChannel) removeSubscriber(subject string, channel chan interface{}) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	if len(e.sub[subject]) == 0 { // Если список подписок пуст
		return
	}
	// Удаление из списка подписок
	subscribers := make([]chan interface{}, len(e.sub[subject])-1)
	i := 0
	for _, subscriber := range e.sub[subject] {
		if channel != subscriber {
			subscribers[i] = subscriber
			i++
		}
	}
	e.sub[subject] = subscribers
	// Закрытие канала по которому доставлялись сообщения
	close(channel)
}

// Функция для публикации сообщения msg для тех кто подписан на subject
func (e *EventChannel) Publish(subject string, msg interface{}) error {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	subscribes, ok := e.sub[subject]
	if !ok { // Если нет подписок на subject
		return nil
	}
	for _, subscribe := range subscribes {
		// Отправка сообщения msg
		// Гарантируется что канал не закрыт!
		go func(channel chan interface{}) {
			channel <- msg
		}(subscribe)
	}
	return nil
}

// Функция для завершения работы системы
func (e *EventChannel) Close(ctx context.Context) error {
	// Если контекст отменен - выходим сразу,
	//работающие хендлеры оставляем работать
	e.mutex.Lock()
	defer e.mutex.Unlock()
	for subject, subscribers := range e.sub {
		for _, subscriber := range subscribers {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				close(subscriber)
			}
		}
		delete(e.sub, subject)
	}
	return nil
}
