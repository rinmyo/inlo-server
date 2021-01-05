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
	tokenDuration = 24 * time.Hour
)

func accessibleRoles() map[string][]string {
	return map[string][]string{
		"adminEvent": {"admin"},
		"/prac.net.StationService/RefreshStation": {"admin", "user"},
		"/prac.net.StationService/InitStation":    {"admin", "user"},
		"/prac.net.StationService/CreateRoute":    {"admin", "user"},
		"/prac.net.StationService/CancelRoute":    {"admin", "user"},
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
	interlockPath := flag.String("interlock", "./resource/interlock.json", "the interlock file path")
	ioPath := flag.String("io", "./resource/io.json", "the io file path")
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

	simulatedController := service.NewSimulatedController(*ioPath)
	var stationController service.StationController = simulatedController
	stationManager := service.NewStationManager(&stationController, *interlockPath)
	stationServer := service.NewStationServer(stationManager)

	go func() {
		for {
			stationManager.RefreshStationStatus()
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
