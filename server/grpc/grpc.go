package grpc

import (
	"context"
	"fmt"
	"log"

	grpcPubsub "github.com/bolotskihUlyanaa/pubsub/protoc/gen/go/pubsub"
	pubsub "github.com/bolotskihUlyanaa/pubsub/server/pubsub"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Реализация gRPC сервиса
type Server struct {
	grpcPubsub.UnimplementedPubSubServer
	channel *pubsub.EventChannel // Шина событий
}

func NewSubServer() *Server {
	return &Server{channel: pubsub.NewEventChannel()}
}

// Функция реализует gRPC эндпоинт Subscribe
func (s *Server) Subscribe(request *grpcPubsub.SubscribeRequest,
	stream grpc.ServerStreamingServer[grpcPubsub.Event]) error {
	key := request.GetKey() // Ключ на который нужно подписаться
	ctx := stream.Context() // Получение контекста из потока
	log.Printf("Subscribe to %s\n", key)
	// Вызов Subscribe на шине событий
	subscribe, err := s.channel.Subscribe(key,
		func(msg interface{}) {
			event := &grpcPubsub.Event{Data: fmt.Sprintf("%v", msg)}
			select {
			case <-ctx.Done():
				log.Printf("Сontext canceled: %s\n", key)
				return
			default:
				// Отправка сообщения клиенту
				if err := stream.Send(event); err != nil {
					log.Printf("Error send message to grpc: %v\n", err)
				}
			}
		},
	)
	if err != nil {
		log.Printf("Error subscribe to %s: %v\n", key, err)
		return status.Error(codes.Internal, "Error event channel")
	}
	<-ctx.Done()            // Ожидание закрытия соединения
	subscribe.Unsubscribe() // Отписка на шине событий
	return status.Error(codes.Canceled, ctx.Err().Error())
}

// Функция реализует gRPC эндпоинт Publish
func (s *Server) Publish(ctx context.Context,
	request *grpcPubsub.PublishRequest) (*emptypb.Empty, error) {
	key := request.GetKey()             // Ключ на который нужно подписаться
	data := request.GetData()           // Сообщение
	err := s.channel.Publish(key, data) // Вызов Publish на шине событий
	if err != nil {
		log.Printf("Error publish %s message %s: %v\n", key, data, err)
		return &emptypb.Empty{},
			status.Error(codes.Internal, "Error event channel")
	}
	log.Printf("Published for %s: %s\n", key, data)
	return &emptypb.Empty{}, nil
}
