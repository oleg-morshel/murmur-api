package grpcserver

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/oleg-morshel/murmur-api/internal/core/domain"
	core_errors "github.com/oleg-morshel/murmur-api/internal/core/errors"
	"github.com/oleg-morshel/murmur-api/pkg/logger"
	"github.com/oleg-morshel/murmur-api/proto/boardpb"
)

type TokenParser interface {
	ParseAccessToken(tokenStr string) (int64, error)
}

type UserProvider interface {
	GetById(ctx context.Context, id int64) (*domain.User, error)
}

type Server struct {
	boardpb.UnimplementedUserServiceServer
	users UserProvider
	log   *logger.Logger
	grpc  *grpc.Server
}

func authInterceptor(tokens TokenParser) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		values := md.Get("authorization")
		if len(values) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization")
		}

		token := strings.TrimPrefix(values[0], "Bearer ")
		if _, err := tokens.ParseAccessToken(token); err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		return handler(ctx, req)
	}
}
func New(users UserProvider, tokens TokenParser, log *logger.Logger) *Server {
	s := &Server{
		users: users,
		log:   log,
		grpc:  grpc.NewServer(grpc.UnaryInterceptor(authInterceptor(tokens))),
	}
	boardpb.RegisterUserServiceServer(s.grpc, s)
	return s
}

func (s *Server) GetUser(ctx context.Context, req *boardpb.GetUserRequest) (*boardpb.GetUserResponse, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id must be positive")
	}

	user, err := s.users.GetById(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, core_errors.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		s.log.Error("grpc: get user failed", slog.Int64("user_id", req.GetId()), slog.Any("error", err))
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &boardpb.GetUserResponse{
		Id:       user.ID,
		Username: user.Username,
	}, nil
}

func (s *Server) Serve(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.log.Info("grpc server listening", slog.String("addr", addr))
	return s.grpc.Serve(lis)
}

func (s *Server) GracefulStop() {
	s.grpc.GracefulStop()
}
