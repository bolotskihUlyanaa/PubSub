package main

import (
	"fmt"
	"log"
	"net"
	"os"

	grpcPubsub "github.com/bolotskihUlyanaa/pubsub/protoc/gen/go/pubsub"
	service "github.com/bolotskihUlyanaa/pubsub/server/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	yaml "gopkg.in/yaml.v2"
)

// Реализация сервера
func main() {
	conf, err := initConfig()
	if err != nil {
		log.Fatal(err)
	}
	service := service.NewSubServer()        // Реализует gRPC сервис
	lis, err := net.Listen("tcp", conf.Port) // Прослушивание входящих соединений
	if err != nil {
		log.Fatalf("Failed to listen %s: %v", conf.Port, err)
	}
	server := grpc.NewServer()  // Создание нового gRPC сервера
	defer server.GracefulStop() // Отложенная остановка сервера
	// Регистрация сервиса service на сервере server
	grpcPubsub.RegisterPubSubServer(server, service)
	reflection.Register(server) // Для grpcurl
	log.Printf("Server listening on %v\n", lis.Addr())
	// Запуск gRPC сервера
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

// Структура конфигурационного файла
type Conf struct {
	Port string `yaml:"port"`
}

// Функция для инициализации конфигурационного файла
func initConfig() (*Conf, error) {
	var cfg Conf
	yamlFile, err := os.ReadFile("conf.yaml")
	if err != nil {
		return nil, fmt.Errorf("Error reading conf file: %v", err)
	}
	err = yaml.Unmarshal(yamlFile, &cfg)
	if err != nil {
		return nil, fmt.Errorf("Error decoding conf file: %v", err)
	}
	return &cfg, nil
}
