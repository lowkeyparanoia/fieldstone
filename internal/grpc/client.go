package grpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/fieldstone/fieldstone/api/proto"
)

// Client provides a convenient wrapper for gRPC client calls
type Client struct {
	conn   *grpc.ClientConn
	client proto.FieldstoneServiceClient
	token  string
}

// ClientConfig holds client configuration
type ClientConfig struct {
	Address    string
	Token      string
	TLSEnabled bool
	CertFile   string
	Timeout    time.Duration
}

// NewClient creates a new gRPC client
func NewClient(cfg ClientConfig) (*Client, error) {
	if cfg.Address == "" {
		cfg.Address = "localhost:50051"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	var opts []grpc.DialOption

	// Configure TLS
	if cfg.TLSEnabled {
		if cfg.CertFile != "" {
			creds, err := credentials.NewClientTLSFromFile(cfg.CertFile, "")
			if err != nil {
				return nil, fmt.Errorf("failed to load TLS cert: %w", err)
			}
			opts = append(opts, grpc.WithTransportCredentials(creds))
		} else {
			opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
		}
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Add auth interceptor if token provided
	if cfg.Token != "" {
		opts = append(opts, grpc.WithUnaryInterceptor(authInterceptor(cfg.Token)))
	}

	conn, err := grpc.Dial(cfg.Address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return &Client{
		conn:   conn,
		client: proto.NewFieldstoneServiceClient(conn),
		token:  cfg.Token,
	}, nil
}

// authInterceptor adds JWT token to requests
func authInterceptor(token string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// Close closes the client connection
func (c *Client) Close() error {
	return c.conn.Close()
}

// Health checks server health
func (c *Client) Health(ctx context.Context) (*proto.HealthResponse, error) {
	return c.client.Health(ctx, &proto.HealthRequest{})
}

// ListCollections lists all collections
func (c *Client) ListCollections(ctx context.Context, req *proto.ListCollectionsRequest) (*proto.ListCollectionsResponse, error) {
	return c.client.ListCollections(ctx, req)
}

// CreateCollection creates a new collection
func (c *Client) CreateCollection(ctx context.Context, req *proto.CreateCollectionRequest) (*proto.Collection, error) {
	return c.client.CreateCollection(ctx, req)
}

// ListRecords lists records in a collection
func (c *Client) ListRecords(ctx context.Context, req *proto.ListRecordsRequest) (*proto.ListRecordsResponse, error) {
	return c.client.ListRecords(ctx, req)
}

// CreateRecord creates a new record
func (c *Client) CreateRecord(ctx context.Context, req *proto.CreateRecordRequest) (*proto.Record, error) {
	return c.client.CreateRecord(ctx, req)
}

// Login authenticates a user
func (c *Client) Login(ctx context.Context, req *proto.LoginRequest) (*proto.AuthResponse, error) {
	resp, err := c.client.Login(ctx, req)
	if err != nil {
		return nil, err
	}
	
	// Update client token after login
	c.token = resp.Token
	return resp, nil
}

// SubscribeRecords subscribes to real-time record updates
func (c *Client) SubscribeRecords(ctx context.Context, req *proto.SubscribeRecordsRequest) (<-chan *proto.RecordEvent, error) {
	stream, err := c.client.SubscribeRecords(ctx, req)
	if err != nil {
		return nil, err
	}

	events := make(chan *proto.RecordEvent)
	go func() {
		defer close(events)
		for {
			event, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				return
			}
			events <- event
		}
	}()

	return events, nil
}

// UploadFile uploads a file via streaming
func (c *Client) UploadFile(ctx context.Context, filename string, contentType string, data []byte) (*proto.File, error) {
	stream, err := c.client.UploadFile(ctx)
	if err != nil {
		return nil, err
	}

	// Send metadata
	if err := stream.Send(&proto.UploadFileRequest{
		Data: &proto.UploadFileRequest_Metadata{
			Metadata: &proto.FileMetadata{
				Filename:    filename,
				ContentType: contentType,
				Size:        int64(len(data)),
			},
		},
	}); err != nil {
		return nil, err
	}

	// Send chunks (4KB chunks)
	chunkSize := 4096
	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		
		if err := stream.Send(&proto.UploadFileRequest{
			Data: &proto.UploadFileRequest_Chunk{
				Chunk: data[i:end],
			},
		}); err != nil {
			return nil, err
		}
	}

	return stream.CloseAndRecv()
}

// Example usage
func ExampleClient() {
	// Create client
	client, err := NewClient(ClientConfig{
		Address:    "localhost:50051",
		TLSEnabled: false,
	})
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// Login
	auth, err := client.Login(context.Background(), &proto.LoginRequest{
		Email:    "user@example.com",
		Password: "password123",
	})
	if err != nil {
		panic(err)
	}
	
	fmt.Println("Token:", auth.Token)

	// Create collection
	collection, err := client.CreateCollection(context.Background(), &proto.CreateCollectionRequest{
		Name: "posts",
		Fields: []*proto.Field{
			{Name: "title", Type: "text", Required: true},
			{Name: "content", Type: "text"},
		},
	})
	if err != nil {
		panic(err)
	}
	
	fmt.Println("Collection created:", collection.Id)

	// List collections
	list, err := client.ListCollections(context.Background(), &proto.ListCollectionsRequest{
		Page:    1,
		PerPage: 30,
	})
	if err != nil {
		panic(err)
	}
	
	fmt.Println("Collections:", len(list.Items))
}
