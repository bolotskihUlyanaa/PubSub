package subpub

// Структура подписчика
type Subscriber struct {
	subpub  *EventChannel    // Шина событий
	subject string           // Имя издателя
	channel chan interface{} // Канал передачи сообщений
}

func NewSubscriber(subpub *EventChannel,
	subject string, channel chan interface{}) *Subscriber {
	return &Subscriber{
		subpub:  subpub,
		subject: subject,
		channel: channel,
	}
}

// Отписка от издателя
func (s *Subscriber) Unsubscribe() {
	s.subpub.removeSubscriber(s.subject, s.channel)
}
