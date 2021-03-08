package producer

import (
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/ychoube/kafka-remote-write/config"
)

type KafkaProducer struct {
	KafkaClientConfig config.KafkaWriterConfig
	KafkaClient       *kafka.Producer
}

func NewKafkaClient(kconfig *config.KafkaWriterConfig) (*KafkaProducer, error) {
	var kp *KafkaProducer
	kp = &KafkaProducer{
		KafkaClientConfig: *kconfig,
	}
	kc, err := kp.NewkafkaProducerClient()
	if err != nil {
		return nil, err
	}
	kp.KafkaClient = kc
	return kp, nil

}

func (k *KafkaProducer) NewkafkaProducerClient() (*kafka.Producer, error) {

	c, err := kafka.NewProducer(k.kafkaClientConfig())
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (k *KafkaProducer) kafkaClientConfig() *kafka.ConfigMap {
	cfg := make(kafka.ConfigMap)
	b := strings.Join(k.KafkaClientConfig.KafkaHosts, ",")
	cfg.SetKey("bootstrap.servers", b)
	cfg.SetKey("compression.codec", "none")
	cfg.SetKey("batch.num.messages", k.KafkaClientConfig.ProducerConfigs.MaxBatchSize)
	cfg.SetKey("go.delivery.reports", false)

	return &cfg
}
