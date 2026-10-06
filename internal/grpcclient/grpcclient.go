package grpcclient

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/oleg-morshel/murmur-api/proto/boardpb"
)

const (
	callTimeout     = 2 * time.Second
	serviceSubject  = "0" // 0 = внутренний сервис, id пользователей начинаются с 1
	serviceTokenTTL = time.Minute
)

type UserClient struct {
	conn   *grpc.ClientConn
	client boardpb.UserServiceClient
}

func NewUserClient(target string, secret []byte) (*UserClient, error) {
	conn, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(authInterceptor(secret)),
	)
	if err != nil {
		return nil, fmt.Errorf("grpc new client: %w", err)
	}

	return &UserClient{
		conn:   conn,
		client: boardpb.NewUserServiceClient(conn),
	}, nil
}

func (c *UserClient) GetUsername(ctx context.Context, id int64) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()

	resp, err := c.client.GetUser(ctx, &boardpb.GetUserRequest{Id: id})
	if err != nil {
		return "", fmt.Errorf("grpc get user %d: %w", id, err)
	}

	return resp.GetUsername(), nil
}

func (c *UserClient) Close() error {
	return c.conn.Close()
}

func newServiceToken(secret []byte) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   serviceSubject,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(serviceTokenTTL)),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
}

func authInterceptor(secret []byte) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		token, err := newServiceToken(secret)
		if err != nil {
			return fmt.Errorf("create service token: %w", err)
		}

		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
