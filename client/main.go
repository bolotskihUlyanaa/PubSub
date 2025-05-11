package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	grpcPubsub "github.com/bolotskihUlyanaa/pubsub/protoc/gen/go/pubsub"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Реализация клиента
func main() {
	// Адрес gRPC передается в качестве аргументов
	if len(os.Args) != 2 {
		log.Fatal("No addres in arguments")
	}
	// Установка соединения с gRPC сервером
	conn, err := grpc.NewClient(os.Args[1],
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Error connect to gRPC server: %v", err)
	}
	defer conn.Close() // Закрытие gRPC соединения
	// Создание gRPC клиента для сервиса pubsub
	client := grpcPubsub.NewPubSubClient(conn)
	for {
		command, err := read() // Чтение команды
		if err != nil {
			log.Println(err)
			continue
		}
		err = runCommand(command, client) // Выполнение команды
		if err != nil {
			log.Println(err)
		}
	}
}

// Функция для выполнения команды клиента
func runCommand(command []string, client grpcPubsub.PubSubClient) error {
	switch command[0] {
	case "sub": // Подписаться на ключ
		if len(command) == 2 {
			go subscribe(command[1], client)
		} else {
			return fmt.Errorf("Incorrect parameters")
		}
	case "pub": // Опубликовать сообщение на ключе
		if len(command) == 3 {
			err := publish(command[1], command[2], client)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("Incorrect parameters")
		}
	default:
		return fmt.Errorf("Command doesnt exist")
	}
	return nil
}

// Функция для чтения команд из стандартного потока ввода
func read() ([]string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	line := scanner.Text()
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("Error scan string: %w", err)
	}
	command := strings.Split(line, " ")
	return command, nil
}

// Функция публикации сообщения на ключе
func publish(key, msg string, client grpcPubsub.PubSubClient) error {
	// Вызов метода Publish gRPC клиента
	_, err := client.Publish(context.Background(),
		&grpcPubsub.PublishRequest{Key: key, Data: msg})
	if err != nil {
		return fmt.Errorf("Error publish from %s: %w", key, err)
	}
	return nil
}

// Функция подписки на ключ
func subscribe(key string, client grpcPubsub.PubSubClient) {
	// Вызов метода Subscribe gRPC клиента
	stream, err := client.Subscribe(context.Background(),
		&grpcPubsub.SubscribeRequest{Key: key})
	if err != nil {
		log.Printf("Error subscribe to %s: %v", key, err)
		return
	}
	// Обработка подписки
	defer func() { // Отложенное закрытие потока
		if err := stream.CloseSend(); err != nil {
			log.Printf("Error close stream for %s:%v", key, err)
		}
	}()
	for {
		msg, err := handlingStream(stream)
		if err != nil {
			log.Printf("Error from %s %v", key, err)
			break
		}
		fmt.Printf("%s : %s\n", key, msg)
	}
	log.Printf("Listening finished %s\n", key)
}

// Функция обработки потока сообщений
func handlingStream(stream grpc.ServerStreamingClient[grpcPubsub.Event]) (
	string, error) {
	// Ожидание получения сообщения или закрытия потока
	event, err := stream.Recv()
	if err == io.EOF { // Закрыт ли поток
		return "", fmt.Errorf("Error stream finished %w", err)
	}
	if err != nil { // Ошибка при получении сообщения
		return "", fmt.Errorf("Error receiving message: %w", err)
	}
	return event.Data, nil
}
