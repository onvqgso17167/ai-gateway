// Package extproc implements the external processing filter handler
// for Envoy's ext_proc gRPC service. It intercepts HTTP requests and
// responses to apply AI gateway logic such as routing, rate limiting,
// and response transformation.
package extproc

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	extprocv3 "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Processor handles the ext_proc bidirectional streaming RPC.
// It processes request and response phases for each HTTP transaction.
type Processor struct {
	extprocv3.UnimplementedExternalProcessorServer
	logger *slog.Logger
}

// NewProcessor creates a new Processor instance with the provided logger.
func NewProcessor(logger *slog.Logger) *Processor {
	return &Processor{
		logger: logger,
	}
}

// Process implements the ExternalProcessorServer interface.
// It handles the bidirectional stream between Envoy and this processor.
func (p *Processor) Process(
	stream extprocv3.ExternalProcessor_ProcessServer,
) error {
	ctx := stream.Context()
	p.logger.InfoContext(ctx, "new ext_proc stream started")

	for {
		select {
		case <-ctx.Done():
			p.logger.InfoContext(ctx, "stream context done", "reason", ctx.Err())
			return nil
		default:
		}

		req, err := stream.Recv()
		if err == io.EOF {
			p.logger.InfoContext(ctx, "stream closed by client")
			return nil
		}
		if err != nil {
			p.logger.ErrorContext(ctx, "error receiving from stream", "error", err)
			return status.Errorf(codes.Internal, "recv error: %v", err)
		}

		resp, err := p.handleRequest(ctx, req)
		if err != nil {
			p.logger.ErrorContext(ctx, "error handling request", "error", err)
			return status.Errorf(codes.Internal, "handle error: %v", err)
		}

		if err := stream.Send(resp); err != nil {
			p.logger.ErrorContext(ctx, "error sending response", "error", err)
			return status.Errorf(codes.Internal, "send error: %v", err)
		}
	}
}

// handleRequest dispatches the incoming ProcessingRequest to the appropriate
// phase handler based on which phase is present in the request.
func (p *Processor) handleRequest(
	ctx context.Context,
	req *extprocv3.ProcessingRequest,
) (*extprocv3.ProcessingResponse, error) {
	switch v := req.Request.(type) {
	case *extprocv3.ProcessingRequest_RequestHeaders:
		return p.handleRequestHeaders(ctx, v.RequestHeaders)
	case *extprocv3.ProcessingRequest_RequestBody:
		return p.handleRequestBody(ctx, v.RequestBody)
	case *extprocv3.ProcessingRequest_ResponseHeaders:
		return p.handleResponseHeaders(ctx, v.ResponseHeaders)
	case *extprocv3.ProcessingRequest_ResponseBody:
		return p.handleResponseBody(ctx, v.ResponseBody)
	default:
		return nil, fmt.Errorf("unknown request type: %T", v)
	}
}

// handleRequestHeaders processes incoming HTTP request headers.
func (p *Processor) handleRequestHeaders(
	ctx context.Context,
	headers *extprocv3.HttpHeaders,
) (*extprocv3.ProcessingResponse, error) {
	p.logger.DebugContext(ctx, "processing request headers")
	return &extprocv3.ProcessingResponse{
		Response: &extprocv3.ProcessingResponse_RequestHeaders{
			RequestHeaders: &extprocv3.HeadersResponse{},
		},
	}, nil
}

// handleRequestBody processes the HTTP request body.
func (p *Processor) handleRequestBody(
	ctx context.Context,
	body *extprocv3.HttpBody,
) (*extprocv3.ProcessingResponse, error) {
	p.logger.DebugContext(ctx, "processing request body", "end_of_stream", body.EndOfStream)
	return &extprocv3.ProcessingResponse{
		Response: &extprocv3.ProcessingResponse_RequestBody{
			RequestBody: &extprocv3.BodyResponse{},
		},
	}, nil
}

// handleResponseHeaders processes incoming HTTP response headers.
func (p *Processor) handleResponseHeaders(
	ctx context.Context,
	headers *extprocv3.HttpHeaders,
) (*extprocv3.ProcessingResponse, error) {
	p.logger.DebugContext(ctx, "processing response headers")
	return &extprocv3.ProcessingResponse{
		Response: &extprocv3.ProcessingResponse_ResponseHeaders{
			ResponseHeaders: &extprocv3.HeadersResponse{},
		},
	}, nil
}

// handleResponseBody processes the HTTP response body.
func (p *Processor) handleResponseBody(
	ctx context.Context,
	body *extprocv3.HttpBody,
) (*extprocv3.ProcessingResponse, error) {
	p.logger.DebugContext(ctx, "processing response body", "end_of_stream", body.EndOfStream)
	return &extprocv3.ProcessingResponse{
		Response: &extprocv3.ProcessingResponse_ResponseBody{
			ResponseBody: &extprocv3.BodyResponse{},
		},
	}, nil
}
