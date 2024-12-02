// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package externalsubscriber

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strings"

	"github.com/ava-labs/avalanchego/utils/logging"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/event"

	pb "github.com/ava-labs/hypersdk/proto/pb/externalsubscriber"
)

var _ event.Subscription[*chain.ExecutedBlock] = (*ExternalSubscriberClient)(nil)

type ExternalSubscriberClient struct {
	conn   *grpc.ClientConn
	client pb.ExternalSubscriberClient
	log    logging.Logger
}

func NewExternalSubscriberClient(
	ctx context.Context,
	log logging.Logger,
	serverAddr string,
	genesisBytes []byte,
) (*ExternalSubscriberClient, error) {
	// Normalize server address by removing "https://" if present
	serverAddr, useTLS := normalizeServerAddress(serverAddr)

	// Setup connection options
	var opts []grpc.DialOption
	if useTLS {
		// Use default TLS credentials
		creds := credentials.NewTLS(&tls.Config{})
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		// Use insecure credentials for plaintext communication
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Establish connection to server
	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial external subscriber server: %w", err)
	}
	client := pb.NewExternalSubscriberClient(conn)

	// Initialize the connection with the external subscriber server
	_, err = client.Initialize(ctx, &pb.InitializeRequest{
		Genesis: genesisBytes,
	})
	if err != nil {
		conn.Close() // Close connection on initialization failure
		return nil, fmt.Errorf("failed to initialize external subscriber client: %w", err)
	}

	log.Debug("connected to external subscriber server", zap.String("address", serverAddr), zap.Bool("useTLS", useTLS))
	return &ExternalSubscriberClient{
		conn:   conn,
		client: client,
		log:    log,
	}, nil
}

func (e *ExternalSubscriberClient) Accept(blk *chain.ExecutedBlock) error {
	blockBytes, err := blk.Marshal()
	if err != nil {
		return err
	}

	req := &pb.BlockRequest{
		BlockData: blockBytes,
	}
	e.log.Debug("sending accepted block to server",
		zap.Stringer("blockID", blk.BlockID),
		zap.Uint64("blockHeight", blk.Block.Hght),
	)
	_, err = e.client.AcceptBlock(context.TODO(), req)
	return err
}

func (e *ExternalSubscriberClient) Close() error {
	return e.conn.Close()
}

// Helper function to determine if TLS should be used and normalize the server address
func normalizeServerAddress(serverAddr string) (string, bool) {
	useTLS := false

	// Remove "https://" prefix if present and set useTLS to true
	if strings.HasPrefix(serverAddr, "https://") {
		serverAddr = strings.TrimPrefix(serverAddr, "https://")
		useTLS = true
	}

	// Check if the port is 443
	host, port, err := net.SplitHostPort(serverAddr)
	if err == nil && port == "443" {
		useTLS = true
		serverAddr = net.JoinHostPort(host, port)
	}

	return serverAddr, useTLS
}
