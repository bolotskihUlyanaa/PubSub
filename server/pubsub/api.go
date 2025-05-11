package subpub

import "context"

// Функция обратного вызова для обработки сообщения,
// которое доставляется подписчикам
type MessageHandler func(msg interface{})

// Интерфейс подписки
type Subscription interface {
	Unsubscribe() // Отказ от подписки
}

// Интерфейс шины событий
type SubPub interface {
	// Создание подписчика на subject
	Subscribe(subject string, cb MessageHandler) (Subscription, error)
	// Публикует сообщение msg для subject
	Publish(subject string, msg interface{}) error
	// Завершение работы системы
	Close(ctx context.Context) error
}

func NewSubPub() SubPub {
	return NewEventChannel()
}
