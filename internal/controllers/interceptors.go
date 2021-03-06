package controllers

import (
	"context"
	"log"
	// // "io"
	"errors"
	"time"
	// "strings"
	//
	// "github.com/google/uuid"
	// "github.com/golang/protobuf/ptypes/empty"
	// // tspb "github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	//
	// pb "github.com/bartmika/mothership-server/proto"
	// "github.com/bartmika/mothership-server/internal/models"
	"github.com/bartmika/mothership-server/internal/utils"
)

func withServerUnaryInterceptor(s *Controller) grpc.ServerOption {
	return grpc.UnaryInterceptor(s.serverInterceptor)
}

// Authorization unary interceptor function to handle authorize per RPC call
func (s *Controller) serverInterceptor(ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	// Skip authorization for the following RPC paths
	ignoreMethods := map[string]bool{
		"/proto.Mothership/Login":        true,
		"/proto.Mothership/Register":     true,
		"/proto.Mothership/RefreshToken": true,
	}

	// Perform our skip on authorization now.
	if ignoreMethods[info.FullMethod] == false {
		sessionUuid, err := s.authorize(ctx)
		if err != nil {
			return nil, err
		}

		// Lookup our user profile in the session or return 500 error.
		user, err := s.manager.GetUser(ctx, sessionUuid)
		if err != nil {
			return nil, err
		}

		// If no user was found then that means our session expired and the
		// user needs to login or use the refresh token.
		if user == nil {
			return nil, errors.New("Session expired - please log in again")
		}

		// Save our user information to the context.
		ctx = context.WithValue(ctx, "user", user)
		ctx = context.WithValue(ctx, "session_uuid", sessionUuid)
	}

	// Calls the handler
	h, err := handler(ctx, req)

	// Logging
	log.Printf("Request - Method:%s\tDuration:%s\tError:%v\n",
		info.FullMethod,
		time.Since(start),
		err)

	return h, err
}

// authorize function authorizes the token received from Metadata
func (s *Controller) authorize(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Errorf(codes.InvalidArgument, "Retrieving metadata is failed")
	}

	authHeader, ok := md["authorization"]
	if !ok {
		return "", status.Errorf(codes.Unauthenticated, "Authorization token is not supplied")
	}

	token := authHeader[0]

	// validateToken function validates the token
	sessionUuid, err := utils.ProcessBearerToken([]byte(s.hmacSecret), token)
	if err != nil {
		return "", status.Errorf(codes.Unauthenticated, err.Error())
	}

	return sessionUuid, nil
}

// DEVELOPERS NOTES:
// - Special thanks to the following tutorial for helping me understand how to
//   implement gRPC authorization:
//   https://shijuvar.medium.com/writing-grpc-interceptors-in-go-bf3e7671fe48
