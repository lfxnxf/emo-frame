package kafka

import (
	"bytes"
	"errors"
	"fmt"
	log "github.com/lfxnxf/emo-frame/logging"
	"io/ioutil"
	stdlog "log"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/samuel/go-zookeeper/zk"
	"go.uber.org/zap"
)

const (
	P_KAFKA_PRE               string = "kfkclient"
	C_KAFKA_PRE               string = "ckfklient"
	KAFKA_INIT                string = "Init"
	KAFKA_GET_PRODUCER_CLIENT string = "PCGet"
	KAFKA_GET_CONSUME_CLIENT  string = "CCGet"
)

const (
	HeadersMessageIDKey = "@m_id"
	HeadersCreateAtKey  = "@m_create"
	HeadersSpanIDKey    = "@m_span"
	HeadersTraceIDKey   = "@m_trace"
)

var (
	KAFKA_CLIENT_NOT_INIT = errors.New("kafka client not init")
	KAFKA_PARAMS_ERROR    = errors.New("kafka params error")
)

var (
	REQUIRED_ACK_NO_RESPONSE    string = "NoResponse"
	REQUIRED_ACK_WAIT_FOR_LOCAL string = "WaitForLocal"
	REQUIRED_ACK_WAIT_FOR_ALL   string = "WaitForAll"
)

var (
	// Logger is kafka client logger
	Logger         = stdlog.New(ioutil.Discard, "[Sarama] ", stdlog.LstdFlags)
	loggerInitOnce = &sync.Once{}
)

type ProducerConfig struct {
	ProducerTo     string `toml:"producer_to"`
	Broken         string `toml:"kafka_broken"`
	RetryMax       int    `toml:"retry_max"`
	RequiredAcks   string `toml:"required_acks"`
	GetError       bool   `toml:"get_error"`
	GetSuccess     bool   `toml:"get_success"`
	RequestTimeout int    `toml:"request_timeout"`
	Printf         bool   `toml:"printf"`
	UseSync        bool   `toml:"use_sync"`
}

type Client struct {
	producer        sarama.AsyncProducer
	conf            ProducerConfig
	perror          chan *ProducerError
	pmessage        chan *ProducerMessage
	headerSupported bool
}

type SyncClient struct {
	producer        sarama.SyncProducer
	conf            ProducerConfig
	headerSupported bool
}

// ProducerMessage is the collection of elements passed to the Producer in order to send a message.
type ProducerMessage struct {
	Topic string // The Kafka topic for this message.
	// The partitioning key for this message. Pre-existing Encoders include
	// StringEncoder and ByteEncoder.
	Key string
	// The actual message to store in Kafka. Pre-existing Encoders include
	// StringEncoder and ByteEncoder.
	Value []byte

	// This field is used to hold arbitrary data you wish to include so it
	// will be available when receiving on the Successes and Errors channels.
	// Sarama completely ignores this field and is only to be used for
	// pass-through data.
	Metadata interface{}

	// Below this point are filled in by the producer as the message is processed

	// Offset is the offset of the message stored on the broker. This is only
	// guaranteed to be defined if the message was successfully delivered and
	// RequiredAcks is not NoResponse.
	Offset int64
	// Partition is the partition that the message was sent to. This is only
	// guaranteed to be defined if the message was successfully delivered.
	Partition int32
	// Timestamp is the timestamp assigned to the message by the broker. This
	// is only guaranteed to be defined if the message was successfully
	// delivered, RequiredAcks is not NoResponse, and the Kafka broker is at
	// least version 0.10.0.
	Timestamp time.Time

	// MessageID
	MessageID string
}

func init() {
	sarama.Logger = Logger
	// https://github.com/Shopify/sarama/issues/959
	sarama.MaxRequestSize = 1000000
}

type logWriter struct {
	l *log.Logger
}

func (l *logWriter) Write(p []byte) (int, error) {
	p = bytes.TrimSpace(p)
	if l.l != nil {
		if bytes.Contains(p, []byte("err")) || bytes.Contains(p, []byte("FAILED")) {
			l.l.Error(string(p))
		} else if bytes.Contains(p, []byte("must")) || bytes.Contains(p, []byte("should")) {
			l.l.Warn(string(p))
		} else {
			l.l.Info(string(p))
		}
	}
	return len(p), nil
}

func initLogger() {
	loggerInitOnce.Do(func() {
		if sarama.Logger == Logger {
			sarama.Logger = stdlog.New(&logWriter{l: &log.Logger{SugaredLogger: log.Log(log.GenLoggerName).SugaredLogger.Desugar().WithOptions(zap.AddCallerSkip(2)).Sugar()}}, "kafka ", 0)
			zk.DefaultLogger = stdlog.New(&logWriter{l: &log.Logger{SugaredLogger: log.Log(log.GenLoggerName).SugaredLogger.Desugar().WithOptions(zap.AddCallerSkip(2)).Sugar()}}, "zookeeper ", 0)
		}
	})
}

// ProducerError is the type of error generated when the producer fails to deliver a message.
// It contains the original ProducerMessage as well as the actual error value.
type ProducerError struct {
	Msg *ProducerMessage
	Err error
}

func makeproducerMsg(spmsg *sarama.ProducerMessage) *ProducerMessage {

	key, _ := spmsg.Key.Encode()
	value, _ := spmsg.Value.Encode()

	return &ProducerMessage{
		Topic:     spmsg.Topic,
		Key:       string(key),
		Value:     value,
		Metadata:  spmsg.Metadata,
		Offset:    spmsg.Offset,
		Partition: spmsg.Partition,
		Timestamp: spmsg.Timestamp,
	}
}

func makeproducerError(sperror *sarama.ProducerError) *ProducerError {

	pm := makeproducerMsg(sperror.Msg)

	return &ProducerError{
		Msg: pm,
		Err: sperror.Err,
	}
}

// getRequiredAcks 如果默认不配置 acks，使用 acks=-1 防止消息丢失
func getRequiredAcks(conf ProducerConfig) (sarama.RequiredAcks, error) {
	if len(conf.RequiredAcks) == 0 {
		conf.RequiredAcks = REQUIRED_ACK_WAIT_FOR_ALL
	}
	if conf.RequiredAcks == REQUIRED_ACK_NO_RESPONSE {
		return sarama.NoResponse, nil
	}
	if conf.RequiredAcks == REQUIRED_ACK_WAIT_FOR_ALL {
		return sarama.WaitForAll, nil
	}
	if conf.RequiredAcks == REQUIRED_ACK_WAIT_FOR_LOCAL {
		return sarama.WaitForLocal, nil
	}
	return 0, KAFKA_PARAMS_ERROR
}

func NewSyncProducerClient(conf ProducerConfig) (*SyncClient, error) {
	initLogger()
	config := sarama.NewConfig()
	acks, err := getRequiredAcks(conf)
	if err != nil {
		return nil, err
	}
	config.Net.KeepAlive = 60 * time.Second
	config.Producer.RequiredAcks = acks
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.MaxMessageBytes = int(sarama.MaxRequestSize - 1) // 1M
	if conf.RequestTimeout == 0 {
		config.Producer.Timeout = 5 * time.Second
	} else {
		config.Producer.Timeout = time.Duration(conf.RequestTimeout) * time.Second
	}
	brokerList := strings.Split(conf.Broken, ",")
	headerSupported := false
	if config.Version.IsAtLeast(sarama.V0_11_0_0) {
		headerSupported = true
	}
	p, err := sarama.NewSyncProducer(brokerList, config)
	if err != nil {
		log.GenLogf("Failed to produce message :%s", err)
		return nil, err
	}

	return &SyncClient{
		producer:        p,
		conf:            conf,
		headerSupported: headerSupported,
	}, nil
}

func NewKafkaClient(conf ProducerConfig) (*Client, error) {
	initLogger()
	log.GenLog("kafka_util,nomal,producer,new kafka client,broken:", conf.Broken, ",productTo:", conf.ProducerTo, ",retryMax:", conf.RetryMax)
	brokerList := strings.Split(conf.Broken, ",")
	config := sarama.NewConfig()
	acks, errf := getRequiredAcks(conf)
	if errf != nil {
		return nil, errf
	}

	config.Net.KeepAlive = 60 * time.Second
	config.Producer.RequiredAcks = acks
	config.Producer.Retry.Max = conf.RetryMax + 1
	config.Producer.Return.Errors = true
	config.Producer.Return.Successes = true
	config.Producer.MaxMessageBytes = int(sarama.MaxRequestSize - 1) // 1M

	headerSupported := false
	if config.Version.IsAtLeast(sarama.V0_11_0_0) {
		headerSupported = true
	}

	producer, err := sarama.NewAsyncProducer(brokerList, config)
	// producer := &p

	if err != nil {
		errf := fmt.Errorf("init syncProcycer error %s", err.Error())
		log.GenLog("kafka_util,error,producer,init error ,broken:", conf.Broken, ",productTo:", conf.ProducerTo, ",retryMax:", conf.RetryMax, ",err:", err.Error())
		return nil, errf
	}

	kc := &Client{
		producer:        producer,
		conf:            conf,
		perror:          make(chan *ProducerError),
		pmessage:        make(chan *ProducerMessage),
		headerSupported: headerSupported,
	}

	go func() {

		errChan := producer.Errors()
		successChan := producer.Successes()
		for {

			select {
			case perr, ok := <-errChan:
				if !ok {
					return
				}
				meta := perr.Msg.Metadata.(*sendMeta)
				if meta.oldMeta != nil || meta.mid != "" {
					log.Warnf("[KafkaProducer] send message to %s error %s, brokers(%s), meta(%v), id(%s)", conf.ProducerTo, perr.Error(), conf.Broken, meta.oldMeta, meta.mid)
				} else {
					log.Warnf("[KafkaProducer] send message to %s error %s, brokers(%s)", conf.ProducerTo, perr.Error(), conf.Broken)
				}
				perr.Msg.Metadata = meta.oldMeta
				if conf.GetError == true {
					kc.perror <- makeproducerError(perr)
				}
			case success, ok := <-successChan:
				if !ok {
					return
				}
				meta := success.Metadata.(*sendMeta)
				success.Metadata = meta.oldMeta
				if conf.GetSuccess == true {
					msg := makeproducerMsg(success)
					kc.pmessage <- msg
				} else {
					if meta.mid != "" || meta.oldMeta != nil {
						log.Infof("send message id %q partition %d, offset %d", meta.mid, success.Partition, success.Offset)
					}
				}
			}
		}
	}()
	return kc, nil
}

func (ksc *Client) Close() error {
	defer func() {
		if err := recover(); err != nil {
			// handler(err)
			log.Errorf("function run panic", err)
		}
	}()
	close(ksc.perror)
	close(ksc.pmessage)
	return (ksc.producer).Close()
}

func (ksc *Client) Errors() <-chan *ProducerError {
	return ksc.perror
}

func (ksc *Client) Success() <-chan *ProducerMessage {
	return ksc.pmessage
}
