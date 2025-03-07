package grpc

import (
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

const grpcSSLPrefix = "ssl://"

func ResolveGrpcCredentials(grpcUrl string) (string, grpc.DialOption) {
	if strings.HasPrefix(grpcUrl, grpcSSLPrefix) {
		addr := strings.TrimPrefix(grpcUrl, grpcSSLPrefix)
		return addr, grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, ""))
	}
	return grpcUrl, grpc.WithTransportCredentials(insecure.NewCredentials())
}
