package proxy

import (
	"context"
	"github.com/lfxnxf/emo-frame/inits"
	"github.com/lfxnxf/emo-frame/resource/kafka"
	"github.com/lfxnxf/emo-frame/resource/kafka/core/consumergroup"
)

type KafkaProducer struct {
	name string
}

type KafkaSyncProducer struct {
	name string
}

type KafkaConsumer struct {
	name string
}

func Topic(ctx context.Context, topic string) string {
	return topic
}

func InitKafkaProducer(name string) *KafkaProducer {
	return &KafkaProducer{name}
}

func InitKafkaSyncProducer(name string) *KafkaSyncProducer {
	return &KafkaSyncProducer{name}
}

func InitKafkaConsumer(name string) *KafkaConsumer {
	return &KafkaConsumer{name}
}

func (k *KafkaSyncProducer) Send(ctx context.Context, message *kafka.ProducerMessage) (int32, int64, error) {
	return inits.SyncProducerClient(k.name).Send(ctx, message)
}

func (k *KafkaSyncProducer) SendSyncMsg(ctx context.Context, topic string, key string, msg []byte) (int32, int64, error) {
	m := &kafka.ProducerMessage{}
	m.Key = key
	m.Topic = Topic(ctx, topic)
	m.Value = msg
	return k.Send(ctx, m)
}

func (k *KafkaProducer) Send(ctx context.Context, message *kafka.ProducerMessage) (int32, int64, error) {
	return inits.KafkaProducerClient(k.name).Send(ctx, message)
}

func (k *KafkaProducer) SendKeyMsg(ctx context.Context, topic string, key string, msg []byte) error {
	m := &kafka.ProducerMessage{}
	m.Key = key
	m.Topic = Topic(ctx, topic)
	m.Value = msg
	_, _, err := k.Send(ctx, m)
	return err
}

// nolint:staticcheck
func (k *KafkaProducer) SendMsg(ctx context.Context, topic string, msg []byte) error {
	m := &kafka.ProducerMessage{
		Topic: topic,
		Key:   "",
		Value: msg,
	}
	_, _, err := inits.KafkaProducerClient(k.name).Send(ctx, m)
	return err
}

func (k *KafkaProducer) Errors(ctx context.Context) <-chan *kafka.ProducerError {
	return inits.KafkaProducerClient(k.name).Errors()
}

func (k *KafkaProducer) Success(ctx context.Context) <-chan *kafka.ProducerMessage {
	return inits.KafkaProducerClient(k.name).Success()
}

func (k *KafkaConsumer) GetClient(ctx context.Context) *consumergroup.ConsumerGroup {
	return inits.KafkaConsumeClient(k.name).GetGroupClient()
}
