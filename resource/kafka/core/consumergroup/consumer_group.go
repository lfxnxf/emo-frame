package consumergroup

import (
	"github.com/Shopify/sarama"
	conf "github.com/lfxnxf/emo-frame/resource/kafka/core/config"
)

// The ConsumerGroup type holds all the information for a consumer that is part
// of a consumer group. Call JoinConsumerGroup to start a consumer.
type ConsumerGroup struct {
	config *conf.Config
	name   string
	group  sarama.ConsumerGroup
	topics []string
}

// Connects to a consumer group, using Zookeeper for auto-discovery
func JoinConsumerGroup(name string, brokers []string, topics []string, config *conf.Config) (cg *ConsumerGroup, err error) {

	if name == "" {
		return nil, sarama.ConfigurationError("Empty consumergroup name")
	}

	if len(topics) == 0 {
		return nil, sarama.ConfigurationError("No topics provided")
	}

	if config == nil {
		config = conf.NewConfig()
	}
	config.Config.ClientID = name

	// Validate configuration
	if err = config.Validate(); err != nil {
		return
	}

	group, err := sarama.NewConsumerGroup(brokers, name, config.Config)
	if err != nil {
		return
	}

	cg = &ConsumerGroup{
		config: config,
		group:  group,
		name:   name,
		topics: topics,
	}
	return
}

func (c *ConsumerGroup) GetGroup() sarama.ConsumerGroup {
	return c.group
}

func (c *ConsumerGroup) GetTopic() []string {
	return c.topics
}
