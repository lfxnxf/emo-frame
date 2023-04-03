package inits

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/lfxnxf/emo-frame/config"
	"github.com/lfxnxf/emo-frame/logging"
	"github.com/lfxnxf/emo-frame/resource/kafka"
	"github.com/lfxnxf/emo-frame/resource/redis"
	"github.com/lfxnxf/emo-frame/resource/sql"
	"github.com/lfxnxf/emo-frame/tpc/inf/go-upstream/registry"
	"github.com/lfxnxf/emo-frame/tpc/inf/go-upstream/upstream"
	"github.com/lfxnxf/emo-frame/trace"
	"github.com/lfxnxf/emo-frame/zd_http/server"
	rpc_client "github.com/lfxnxf/emo-frame/zd_rpc/client"
	"path/filepath"
	"strings"
	"sync"
)

const (
	LogRotateHour  = "hour"
	LogRotateDay   = "day"
	LogRotateMonth = "month"
)

type Option func(*Default)

var (
	_default       = new(Default)
	ServiceManager = registry.NewServiceManager(logging.Log(logging.GenLoggerName))
)

type Default struct {
	once            *sync.Once
	configPath      string
	configInstance  config.Instance
	config          config.Config
	logDir          string
	logLevel        string
	logRotate       string
	mysqlClients    sync.Map
	redisClients    sync.Map
	consumeClients  sync.Map
	producerClients sync.Map
	rpcClients      sync.Map
}

func ConfigPath(configPath string) Option {
	return func(d *Default) {
		d.configPath = configPath
	}
}

func Once() Option {
	return func(d *Default) {
		d.once = new(sync.Once)
	}
}

func ConfigInstance(configInstance config.Instance) Option {
	return func(d *Default) {
		d.configInstance = configInstance
	}
}

func Init(opts ...Option) {
	_default.Start(opts...)
}

func (d *Default) Start(opts ...Option) {
	for _, opt := range opts {
		opt(d)
	}
	// config
	err := d.loadLocalConfig()
	if err != nil {
		panic(err)
	}

	d.once.Do(func() {
		// log
		d.initLogger()
		// trace
		d.initTrace()

		// mysql
		if len(d.config.Database) > 0 {
			err := d.initSqlClient(d.config.Database)
			if err != nil {
				panic(err)
			}
		}

		// redis
		if len(d.config.Redis) > 0 {
			err := d.initRedisClient(d.config.Redis)
			if err != nil {
				panic(err)
			}
		}

		// kafka producer
		if len(d.config.KafkaProducer) > 0 {
			err := d.initKafkaProducer(d.config.KafkaProducer)
			if err != nil {
				panic(err)
			}
		}

		// kafka consumer
		if len(d.config.KafkaConsumer) > 0 {
			err := d.initKafkaConsume(d.config.KafkaConsumer)
			if err != nil {
				panic(err)
			}
		}

		// rpc client
		if len(d.config.RpcClient) > 0 {
			d.initRpcClient(d.config.RpcClient)
		}

	})
}

func (d *Default) loadLocalConfig() error {
	if len(d.configPath) == 0 {
		return nil
	}
	if _, err := toml.DecodeFile(d.configPath, d.configInstance); err != nil {
		panic(err)
	}
	d.config = d.configInstance.GetBase()
	return nil
}

func (d *Default) initLogger() {
	if len(d.config.Log.LogPath) == 0 {
		d.config.Log.LogPath = "logs"
	}
	d.logDir = d.config.Log.LogPath

	// Init common logger
	logging.InitCommonLog(logging.CommonLogConfig{
		Pathprefix:      d.config.Log.LogPath,
		Rotate:          d.config.Log.Rotate,
		GenLogLevel:     d.config.Log.GenLogLevel,
		BalanceLogLevel: d.config.Log.BalanceLogLevel,
	})

	// upstream logger
	upstream.SetLogger(logging.Log(logging.BalanceLoggerName))

	ServiceManager.SetLogger(logging.Log(logging.BalanceLoggerName))

	if d.config.Log.Rotate == LogRotateDay {
		logging.SetRotateByDay()
	} else {
		logging.SetRotateByHour()
	}
	if len(d.config.Log.Level) > 0 {
		logging.SetLevelByString(d.config.Log.Level)
	} else {
		logging.SetLevelByString(d.logLevel)
	}
	// will init debug info error logger inside
	logging.SetOutputPath(d.logDir)

	// internal logger
	rotateType := d.config.Log.Rotate

	// 调用日志和请求日志
	var accessLog *logging.Logger
	if !d.config.Log.AccessLogOff {
		accessLog = logging.NewLogging(filepath.Join(d.logDir, "access.log"))
		if rotateType == "day" {
			accessLog.SetRotateByDay()
		}
	}

	infoLog := logging.NewLogging(filepath.Join(d.logDir, "info.log"))
	if rotateType == "day" {
		infoLog.SetRotateByDay()
	}

	debugLog := logging.NewLogging(filepath.Join(d.logDir, "debug.log"))
	if rotateType == "day" {
		infoLog.SetRotateByDay()
	}

	// 错误日志
	errorLog := logging.NewLogging(filepath.Join(d.logDir, "error.log"))
	errorLog.SetLevelByString("error")
	if rotateType == "day" {
		errorLog.SetRotateByDay()
	}

	// sql日志
	sqlLog := logging.NewLogging(filepath.Join(d.logDir, "sql.log"))
	if rotateType == "day" {
		sqlLog.SetRotateByDay()
	}

	// 请求下游business日志
	businessLog := logging.NewLogging(filepath.Join(d.logDir, "business.log"))
	if rotateType == "day" {
		businessLog.SetRotateByDay()
	}

	if logging.DefaultKit == nil {
		logging.DefaultKit = logging.NewKit(accessLog, errorLog, infoLog, debugLog, sqlLog, businessLog)
	}

	go logging.RemoveExpireLogs(d.logDir, d.config.Log.StorageDay)
}

func (d *Default) initTrace() {
	if len(d.config.Telemetry.Name) == 0 {
		d.config.Telemetry.Name = d.config.Server.ServiceName
	}
	trace.StartAgent(d.config.Telemetry)
}

func (d *Default) initSqlClient(sqlList []sql.GroupConfig) error {
	for _, c := range sqlList {
		if _, ok := d.mysqlClients.Load(c.Name); ok {
			continue
		}
		if len(c.LogLevel) == 0 {
			c.LogLevel = strings.ToLower(d.config.Log.Level)
		}
		g, err := sql.NewGroup(c)
		if err != nil {
			return err
		}
		_ = sql.SQLGroupManager.Add(c.Name, g)
		d.mysqlClients.LoadOrStore(c.Name, g)
	}
	return nil
}

func (d *Default) initRedisClient(redisList []redis.Conf) error {
	for _, c := range redisList {
		if _, ok := d.redisClients.Load(c.Name); ok {
			continue
		}
		r := c.NewRedis()
		d.redisClients.LoadOrStore(c.Name, r)
	}
	return nil
}

func (d *Default) initRpcClient(rpcConf []rpc_client.RpcClientConf) {
	for _, c := range rpcConf {
		d.rpcClients.LoadOrStore(c.Name, rpc_client.NewRpcClient(c))
	}
}

func (d *Default) KafkaConsumeClient(consumeFrom string) *kafka.ConsumeClient {
	if client, ok := d.consumeClients.Load(consumeFrom); ok {
		if v, ok1 := client.(*kafka.ConsumeClient); ok1 {
			return v
		}
	}
	fmt.Printf("namespace %s kafka consume client %s not exist\n", consumeFrom)
	logging.GenLogf("namespace %s kafka consume client %s not exist", consumeFrom)
	return nil
}

func (d *Default) KafkaProducerClient(producerTo string) *kafka.Client {
	if client, ok := d.producerClients.Load(producerTo); ok {
		if v, ok := client.(*kafka.Client); ok {
			return v
		}
		fmt.Printf("kafka producer %s type not match, should use SyncProducerClient()\n", producerTo)
		logging.GenLogf("kafka producer %s type not match, should use SyncProducerClient()", producerTo)
		return nil
	}
	fmt.Printf("kafka producer client %s to not exist\n", producerTo)
	logging.GenLogf("kafka producer client %s to not exist", producerTo)
	return nil
}

func (d *Default) SyncProducerClient(producerTo string) *kafka.SyncClient {
	if client, ok := d.producerClients.Load(producerTo); ok {
		if v, ok := client.(*kafka.SyncClient); ok {
			return v
		}
		fmt.Printf("namespace %s kafka sync producer %s type not match, should use KafkaProducerClient()\n", producerTo)
		logging.GenLogf("namespace %s kafka sync producer %s type not match, should use KafkaProducerClient()", producerTo)
		return nil
	}
	fmt.Printf("kafka sync producer client %s not exist\n", producerTo)
	logging.GenLogf("kafka sync producer client %s not exist", producerTo)
	return nil
}

func (d *Default) initKafkaProducer(kpcList []kafka.ProducerConfig) error {
	for _, item := range kpcList {
		if _, ok := d.producerClients.Load(item.ProducerTo); ok {
			continue
		}
		if item.UseSync {
			client, err := kafka.NewSyncProducerClient(item)
			if err != nil {
				return err
			}
			// 忽略已存在的记录
			d.producerClients.LoadOrStore(item.ProducerTo, client)
		} else {
			client, err := kafka.NewKafkaClient(item)
			if err != nil {
				return err
			}
			d.producerClients.LoadOrStore(item.ProducerTo, client)
		}
	}
	return nil
}

func (d *Default) initKafkaConsume(kccList []kafka.ConsumeConfig) error {
	for _, item := range kccList {
		if _, ok := d.consumeClients.Load(item.ConsumeFrom); ok {
			continue
		}
		client, err := kafka.NewKafkaConsumeClient(item)
		if err != nil {
			return err
		}
		d.consumeClients.LoadOrStore(item.ConsumeFrom, client)
	}
	return nil
}

func NewHttpServer() *server.HttpServer {
	return server.NewHttpServer(_default.config.Server)
}

func SQLClient(name string) *sql.Group {
	if client, ok := _default.mysqlClients.Load(name); ok {
		if v, ok1 := client.(*sql.Group); ok1 {
			return v
		}
	}
	return nil
}

func SyncProducerClient(producerTo string) *kafka.SyncClient {
	return _default.SyncProducerClient(producerTo)
}

func KafkaProducerClient(producerTo string) *kafka.Client {
	return _default.KafkaProducerClient(producerTo)
}

func KafkaConsumeClient(message string) *kafka.ConsumeClient {
	return _default.KafkaConsumeClient(message)
}

func RedisClient(name string) *redis.Redis {
	if client, ok := _default.redisClients.Load(name); ok {
		if v, ok1 := client.(*redis.Redis); ok1 {
			return v
		}
	}
	return nil
}

func RpcClient(name string) *rpc_client.RpcClient {
	if client, ok := _default.rpcClients.Load(name); ok {
		if v, ok1 := client.(*rpc_client.RpcClient); ok1 {
			return v
		}
	}
	return nil
}

func SetRpcClient(name, address string) *rpc_client.RpcClient {
	v, ok := _default.rpcClients.Load(name)
	var client *rpc_client.RpcClient
	if ok {
		client, _ = v.(*rpc_client.RpcClient)
	} else {
		client = rpc_client.NewRpcClient(rpc_client.RpcClientConf{
			Name:    name,
			Address: address,
		})
		_default.rpcClients.LoadOrStore(name, client)
	}
	return client
}

func GetConfigInstance() config.Instance {
	return _default.configInstance
}
