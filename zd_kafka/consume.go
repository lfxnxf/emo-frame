package zd_kafka

import (
	"context"
	"errors"
	"github.com/Shopify/sarama"
	"github.com/lfxnxf/emo-frame/inits/proxy"
	"github.com/lfxnxf/emo-frame/logging"
	"github.com/lfxnxf/emo-frame/resource/kafka"
	"github.com/lfxnxf/emo-frame/trace"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"sync"
	"syscall"
)

type consumer struct {
	client *proxy.KafkaConsumer
	worker int
	p      processor
}

type Consumers struct {
	consumers    []consumer
	consumerChan chan consumer
	started      chan struct{}
	initQuited   chan struct{}

	quited    *atomic.Bool
	isStarted bool
	lock      sync.Mutex
}

type WaitGroupWrapper struct {
	sync.WaitGroup
}

func (w *WaitGroupWrapper) Go(cb func()) {
	w.Add(1)
	go func() {
		cb()
		w.Done()
	}()
}

type processor interface {
	Process(context.Context, string, []byte) error
}

type ProcFunc func(context.Context, string, []byte) error

func (f ProcFunc) Process(ctx context.Context, msgId string, val []byte) error {
	return f(ctx, msgId, val)
}

func NewConsumers() *Consumers {
	cs := &Consumers{
		consumerChan: make(chan consumer),
		started:      make(chan struct{}),
		initQuited:   make(chan struct{}),
		quited:       atomic.NewBool(false),
	}
	cs.init()
	return cs
}

func (cs *Consumers) init() {
	go func() {
		defer close(cs.initQuited)
		for {
			select {
			case c, ok := <-cs.consumerChan:
				if !ok {
					return
				}
				cs.consumers = append(cs.consumers, c)
			case <-cs.started:
				return
			}
		}
	}()
}

func (cs *Consumers) AddConsumer(client *proxy.KafkaConsumer, workerSize int, p processor) error {
	if cs.quited.Load() {
		return nil
	}

	select {
	case cs.consumerChan <- consumer{
		client: client,
		worker: workerSize,
		p:      p,
	}:
	}

	return nil
}

// Start block until context Done or SIGINT/SIGTERM received
func (cs *Consumers) Start(ctx context.Context) error {
	var wg WaitGroupWrapper
	sigCtx := WithSignals(ctx, syscall.SIGINT, syscall.SIGTERM)

	if !cs.isStarted {
		cs.lock.Lock()
		if !cs.isStarted {
			cs.isStarted = true
			close(cs.started)
		}
		cs.lock.Unlock()
	}
	<-cs.initQuited

	for _, c := range cs.consumers {
		c := c
		wg.Go(func() { c.consume(sigCtx) })
	}

	wg.Go(func() {
		for {
			select {
			case <-sigCtx.Done():
				cs.quit()
				return
			case c, ok := <-cs.consumerChan:
				if !ok {
					return
				}
				wg.Go(func() { c.consume(sigCtx) })
			}
		}
	})

	wg.Wait()
	return sigCtx.Err()
}

func (cs *Consumers) quit() {
	cs.quited.Store(true)

	// drain channel's element
	for {
		select {
		case <-cs.consumerChan:
		default:
			goto DONE
		}
	}

DONE:
	return
}

var (
	clientNotFound = errors.New("get client failed")
)

func (c *consumer) consume(ctx context.Context) error {
	log := logging.For(ctx, "func", "consumer.consume")

	client := c.client
	if client == nil {
		log.Errorw("get kafka client clientIsNil")
		return clientNotFound
	}

	var wg WaitGroupWrapper
	group := client.GetClient(ctx)
	for i := 0; i < c.worker; i++ {
		wg.Go(func() {
			for {
				err := group.GetGroup().Consume(ctx, group.GetTopic(), consumerGroupHandler{
					p: c.p,
				})
				if err != nil {
					logging.Errorw("consume error",
						zap.Error(err),
						zap.Any("topics", group.GetTopic()),
					)
					return
				}
			}
		})
	}
	wg.Wait()
	return nil
}

var (
	global *Consumers
	once   sync.Once
)

func initGlobal() {
	once.Do(func() {
		global = NewConsumers()
	})
}

func AddConsumer(client *proxy.KafkaConsumer, workerSize int, p processor) error {
	initGlobal()
	return global.AddConsumer(client, workerSize, p)
}

func StartConsumers(ctx context.Context) error {
	initGlobal()
	return global.Start(ctx)
}

type consumerGroupHandler struct {
	p processor
}

func (consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h consumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		carrier := make(map[string]string)
		ctx := context.Background()
		for _, h := range msg.Headers {
			carrier[string(h.Key)] = string(h.Value)
		}
		ctx = trace.GetContext(ctx, carrier[kafka.HeadersTraceIDKey], carrier[kafka.HeadersSpanIDKey])
		err := h.p.Process(ctx, carrier[kafka.HeadersMessageIDKey], msg.Value)
		if err != nil {
			logging.Errorw("consume error",
				zap.Error(err),
				zap.Any("msg", msg),
			)
			return err
		}
		sess.MarkMessage(msg, "")
	}
	return nil
}
