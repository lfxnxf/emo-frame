package rpc_client

import (
	"github.com/lfxnxf/emo-frame/logging"
	"github.com/lfxnxf/emo-frame/tools/syncx"
	"github.com/lfxnxf/emo-frame/zd_rpc/middleware"
	"google.golang.org/grpc"
)

type RpcClient struct {
	conf         RpcClientConf
	conn         *grpc.ClientConn
	singleFlight syncx.SingleFlight
}

func NewRpcClient(c RpcClientConf) *RpcClient {
	return &RpcClient{
		conf:         c,
		singleFlight: syncx.NewSingleFlight(),
	}
}

type RpcClientConf struct {
	Name    string `toml:"name"`
	Address string `toml:"address"`
}

func (c *RpcClient) GetRpcConn(options ...grpc.DialOption) *grpc.ClientConn {
	if c.conn != nil {
		return c.conn
	}
	_, _ = c.singleFlight.Do(c.conf.Name, func() (interface{}, error) {
		opt := middleware.GetClientOpts()
		opt = append(opt, options...)
		conn, err := grpc.Dial(c.conf.Address, opt...)
		if err != nil {
			logging.Fatalf("did not connect: %v", err)
			return nil, err
		}
		c.conn = conn
		return conn, nil
	})
	return c.conn
}

func (c *RpcClient) Close() {
	_ = c.conn.Close()
}
