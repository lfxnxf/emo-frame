package middleware

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func GetServerOpts() []grpc.UnaryServerInterceptor {
	// todo max_connects
	// todo prometheus
	return []grpc.UnaryServerInterceptor{
		loggingAccess, // 生成access_log
		getTrace,      // 设置trace
		recoverSysMW,  // recover
	}
}

func GetClientOpts() []grpc.DialOption {
	// todo max_connects
	// todo prometheus
	return []grpc.DialOption{
		setTrace(), // 设置trace
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
}
