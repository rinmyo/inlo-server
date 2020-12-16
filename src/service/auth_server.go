package service

import (
	"context"
	log "github.com/sirupsen/logrus"
	"pracserver/src/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthServer is the server for authentication
type AuthServer struct {
	userCollection *UserCollection
	jwtManager     *JWTManager
}

// NewAuthServer returns a new auth server
func NewAuthServer(userCollection *UserCollection, jwtManager *JWTManager) *AuthServer {
	return &AuthServer{userCollection, jwtManager}
}

// Login is a unary RPC to login user
func (server *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, err := server.userCollection.Find(req.GetUsername())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "cannot find user: %v", err)
	}

	if user == nil || !user.IsCorrectPassword(req.GetPassword()) {
		return nil, status.Errorf(codes.Aborted, "incorrect password")
	}

	token, err := server.jwtManager.Generate(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot generate access token")
	}

	res := &pb.LoginResponse{AccessToken: token}
	log.Info("give token: ", res.AccessToken)
	return res, nil
}
