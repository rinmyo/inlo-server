package main

import (
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"pracserver/src/config"
	"pracserver/src/pb"
	"pracserver/src/service"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func createUser(userCollection service.UserCollection, username, password, role string) error {
	user, err := service.NewUser(username, password, role)
	if err != nil {
		return err
	}
	return userCollection.Save(user)
}

const (
	secretKey     = "secret"
	tokenDuration = 100 * time.Second
)

func accessibleRoles() map[string][]string {
	return map[string][]string{
		"adminEvent": {"admin"},
		"userEvent":  {"admin", "user"},
	}
}

func runGRPCServer(
	authServer pb.AuthServiceServer,
	stationServer pb.StationServiceServer,
	jwtManager *service.JWTManager,
	listener net.Listener,
) error {
	interceptor := service.NewAuthInterceptor(jwtManager, accessibleRoles())
	serverOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(interceptor.Unary()),
		grpc.StreamInterceptor(interceptor.Stream()),
	}

	grpcServer := grpc.NewServer(serverOptions...)

	pb.RegisterAuthServiceServer(grpcServer, authServer)
	pb.RegisterStationServiceServer(grpcServer, stationServer)
	reflection.Register(grpcServer)

	log.Printf("Start GRPC server at %s", listener.Addr().String())
	return grpcServer.Serve(listener)
}

const logFormat = "log_20060102150405"

func main() {
	port := flag.Int("port", 8080, "the server port")
	flag.Parse()

	client, disconnectMongoDB, err := service.NewMongoClient("localhost")
	if err != nil {
		log.Fatal("cannot make a client: ", err)
	}
	//setting the logger
	logDb := client.Database("pracserver_log")
	logCollection := service.NewLogCollection(logDb, logFormat)
	hook := service.NewHookerFromCollection(logCollection)
	log.AddHook(hook)
	//setting users
	serverDb := client.Database("pracserver")
	userCollection := service.NewUserCollection(serverDb)
	err = createUser(*userCollection, config.UserId(), config.UserPassword(), "admin")
	if err != nil {
		log.Fatal("cannot seed users: ", err)
	}

	jwtManager := service.NewJWTManager(secretKey, tokenDuration)
	authServer := service.NewAuthServer(userCollection, jwtManager)

	simulatedController := service.NewSimulatedController()
	var stationController service.StationController = simulatedController
	stationManager := service.NewStationManager(&stationController, "./resource/interlock.json")
	stationServer := service.NewStationServer(stationManager)
	stationManager.RefreshStationStatus()

	go func() {
		for {
			time.Sleep(1 * time.Second)
			simulatedController.UpdateTurnoutStatus("3", pb.Turnout_REVERSED)
			time.Sleep(1 * time.Second)
			simulatedController.UpdateTurnoutStatus("3", pb.Turnout_NORMAL)
		}
	}()

	address := fmt.Sprintf("0.0.0.0:%d", *port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("cannot start server: ", err)
	}

	err = runGRPCServer(authServer, stationServer, jwtManager, listener) //here program blocking
	if err != nil {
		log.Fatal("cannot start server: ", err)
	}

	err = disconnectMongoDB()
	if err != nil {
		log.Fatal("cannot close connect", err)
	}
}
