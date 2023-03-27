package kafka

import (
	"context"
	"github.com/Shopify/sarama"
	"github.com/lfxnxf/emo-frame/resource/kafka/core/config"
	"github.com/lfxnxf/emo-frame/resource/kafka/core/consumergroup"
	"github.com/lfxnxf/emo-frame/tpc/inf/go-tls"
	"os"
	"strings"
	"sync"
	"time"
)

type ConsumeConfig struct {
	ConsumeFrom    string `toml:"consume_from"`
	KafkaBroken    string `toml:"kafka_broken"`
	Topic          string `toml:"topic"`
	Group          string `toml:"group"`
	InitOffset     int64  `toml:"init_offset"`
	ProcessTimeout int    `toml:"process_timeout"`
	CommitInterval int    `toml:"commit_interval"`
	GetError       bool   `toml:"get_error"`
	TraceEnable    bool   `toml:"trace_enable"`
	ConsumeAll     bool   `toml:"consume_all"`
}

type ConsumeClient struct {
	consumer    *consumergroup.ConsumerGroup
	conf        ConsumeConfig
	err         chan error
	closeChan   chan bool
	messageChan chan *ConsumerMessage
	mu          sync.Mutex
}

type RecordHeader struct {
	Key   []byte
	Value []byte
}

type ConsumerMessage struct {
	Key, Value     []byte
	Topic          string
	Partition      int32
	Offset         int64
	Timestamp      time.Time // only set if kafka is version 0.10+, inner message timestamp
	BlockTimestamp time.Time // only set if kafka is version 0.10+, outer (compressed) block timestamp
	MessageID      string
	CreateAt       time.Time
	headers        []*RecordHeader // only set if kafka is version 0.11+
	ctx            context.Context
}

func (m *ConsumerMessage) Context() context.Context {
	tls.SetContext(m.ctx)
	return m.ctx
}

type ConsumeCallback interface {
	Process(values []byte)
}

func NewKafkaConsumeClient(conf ConsumeConfig) (*ConsumeClient, error) {
	initLogger()

	cfg := config.NewConfig()
	version, err := sarama.ParseKafkaVersion("2.1.1")
	if err != nil {
		panic(err)
	}

	cfg.Config.Version = version
	cfg.Config.Net.KeepAlive = 5 * time.Second

	cfg.Config.Consumer.Offsets.Initial = conf.InitOffset
	cfg.Config.Consumer.Offsets.AutoCommit.Enable = true
	cfg.Config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategySticky

	kafkaTopics := strings.Split(conf.Topic, ",")
	if conf.ConsumeAll {
		hostname, err := os.Hostname()
		if err != nil {
			panic("get hostname error")
		}
		conf.Group = conf.Group + "_" + hostname
	}
	consumer, consumerErr := consumergroup.JoinConsumerGroup(conf.Group, strings.Split(conf.KafkaBroken, ","), kafkaTopics, cfg)
	if consumerErr != nil {
		return nil, consumerErr
	}

	kcc := &ConsumeClient{
		consumer:    consumer,
		conf:        conf,
		err:         make(chan error),
		messageChan: nil,
	}
	return kcc, nil
}

func (kcc *ConsumeClient) GetGroupClient() *consumergroup.ConsumerGroup {
	return kcc.consumer
}

func (kcc *ConsumeClient) Close() error {
	kcc.mu.Lock()
	defer kcc.mu.Unlock()
	if kcc.closeChan != nil {
		close(kcc.closeChan)
		kcc.closeChan = nil
	}
	return nil
}

func (kcc *ConsumeClient) Errors() <-chan error {
	return kcc.err
}
