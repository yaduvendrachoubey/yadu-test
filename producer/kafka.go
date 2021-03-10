package producer

import (

	//"github.com/confluentinc/confluent-kafka-go/kafka"
	kafka "github.com/segmentio/kafka-go"

	"github.com/ychoube/kafka-remote-write/config"
)

type KafkaMessages struct {
	MessageList []kafka.Message
}

type KafkaProducer struct {
	KafkaClientConfig config.KafkaWriterConfig
	KafkaClient       *kafka.Writer
}

func NewKafkaClient(kconfig *config.KafkaWriterConfig) (*KafkaProducer, error) {
	var kp *KafkaProducer
	kp = &KafkaProducer{
		KafkaClientConfig: *kconfig,
	}
	kp.KafkaClient = kp.NewkafkaProducerClient()
	return kp, nil
}

func (k *KafkaProducer) NewkafkaProducerClient() *kafka.Writer {

	return &kafka.Writer{
		Addr:        kafka.TCP(k.KafkaClientConfig.KafkaHosts...),
		Topic:       k.KafkaClientConfig.TopicName,
		MaxAttempts: 3,
	}
}
