package config

import (
	"github.com/lfxnxf/emo-frame/resource/kafka"
	"github.com/lfxnxf/emo-frame/resource/redis"
	"github.com/lfxnxf/emo-frame/resource/sql"
	"github.com/lfxnxf/emo-frame/trace"
	"github.com/lfxnxf/emo-frame/zd_http/server"
	rpc_client "github.com/lfxnxf/emo-frame/zd_rpc/client"
	rpc_server "github.com/lfxnxf/emo-frame/zd_rpc/server"
)

type Instance interface {
	GetBase() Config
}

type Config struct {
	Log           Log                        `toml:"log"`
	Server        server.HttpServerConfig    `toml:"server"`
	RpcServer     rpc_server.RpcServerConfig `toml:"rpc_server"`
	RpcClient     []rpc_client.RpcClientConf `toml:"rpc_client"`
	Telemetry     trace.Config               `toml:"telemetry"`
	Database      []sql.GroupConfig          `toml:"database"`
	Redis         []redis.Conf               `toml:"redis"`
	KafkaProducer []kafka.ProducerConfig     `toml:"kafka_producer_client"`
	KafkaConsumer []kafka.ConsumeConfig      `toml:"kafka_consume"`
}

type Log struct {
	Level              string `toml:"level"`
	Rotate             string `toml:"rotate"`
	AccessRotate       string `toml:"access_rotate"`
	AccessLog          string `toml:"access_log"`
	BusinessLog        string `toml:"business_log"`
	ServerLog          string `toml:"server_log"`
	StatLog            string `toml:"stat_log"`
	ErrorLog           string `toml:"err_log"`
	LogPath            string `toml:"log_path"`
	BalanceLogLevel    string `toml:"balance_log_level"`
	GenLogLevel        string `toml:"gen_log_level"`
	AccessLogOff       bool   `toml:"access_log_off"`
	BusinessLogOff     bool   `toml:"business_log_off"`
	RequestBodyLogOff  bool   `toml:"request_log_off"`
	RespBodyLogMaxSize int    `toml:"response_log_max_size"` // -1:不限制;默认1024字节;
	SuccessStatCode    []int  `toml:"success_stat_code"`
	StorageDay         int64  `toml:"storage_day"` // 日志保留时间
}
