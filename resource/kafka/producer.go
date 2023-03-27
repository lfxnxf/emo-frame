package kafka

import (
	"context"
	"fmt"
	"github.com/lfxnxf/emo-frame/trace"
	"github.com/lfxnxf/emo-frame/utils"
	"time"

	"github.com/Shopify/sarama"
	uuid "github.com/satori/go.uuid"
)

type nsKeyType struct{}

var nsKey = nsKeyType{}

func WithNSKey(ctx context.Context, ns string) context.Context {
	return context.WithValue(ctx, nsKey, ns)
}

func NSKey(ctx context.Context) (string, bool) {
	ns, ok := ctx.Value(nsKey).(string)
	return ns, ok
}

type SendResponse struct {
	Partition int32
	Offset    int64
	Err       error
}

type sendMeta struct {
	eventTime time.Time
	mid       string
	oldMeta   interface{}
}

// Send message to kafka cluster, ctx is http/rpc context
// headers rfs = https://cwiki.apache.org/confluence/display/KAFKA/KIP-82+-+Add+Record+Headers
func (ksc *SyncClient) Send(ctx context.Context, message *ProducerMessage) (int32, int64, error) {
	msg, _ := generateMessageSpan(ctx, message)
	partition, offset, err := ksc.producer.SendMessage(msg)
	return partition, offset, err
}

func (ksc *Client) Send(ctx context.Context, message *ProducerMessage) (int32, int64, error) {
	msg, now := generateMessageSpan(ctx, message)
	msg.Metadata = &sendMeta{eventTime: now, mid: message.MessageID, oldMeta: message.Metadata}
	(ksc.producer).Input() <- msg
	return -1, -1, nil
}

func (ksc *SyncClient) Close() error {
	return ksc.producer.Close()
}

func generateMessageSpan(ctx context.Context, message *ProducerMessage) (*sarama.ProducerMessage, time.Time) {
	msg := &sarama.ProducerMessage{}
	now := time.Now()

	carrier := make(map[string]string)
	if message.MessageID == "" {
		message.MessageID = utils.GetSeqID()
	}

	span, ok := ctx.Value(string(trace.CtxTraceSpanKey)).(trace.Span)
	if ok {
		carrier[HeadersTraceIDKey] = span.Trace()
		carrier[HeadersSpanIDKey] = span.Span()
	}

	carrier[HeadersMessageIDKey] = message.MessageID
	carrier[HeadersCreateAtKey] = fmt.Sprintf("%d", now.UnixNano())

	headers := make([]sarama.RecordHeader, 0, len(carrier))
	for k, v := range carrier {
		headers = append(headers,
			sarama.RecordHeader{
				Key:   []byte(k),
				Value: []byte(v),
			},
		)
	}
	msg.Headers = headers

	msg.Topic = message.Topic
	if message.Partition <= 0 {
		msg.Partition = int32(-1)
	}
	if message.Key == "" {
		u1 := uuid.NewV4()
		msg.Key = sarama.ByteEncoder(u1.Bytes())
	} else {
		msg.Key = sarama.StringEncoder(message.Key)
	}
	msg.Value = sarama.ByteEncoder(message.Value)
	if message.Timestamp.IsZero() {
		msg.Timestamp = now
	} else {
		msg.Timestamp = message.Timestamp
	}
	return msg, now
}
