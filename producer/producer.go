package producer

import (
	"github.com/ychoube/kafka-remote-write/config"
)


type Producers struct {
	ProducersConfig      config.RemoteWriteConfig
	KafkaProducers       []*KafkaProducer
	GoogleCloudProducers []*GoogleCloudProducer
}

func (p *Producers) ProducerClients() error {
	// read config
	for _, k := range p.ProducersConfig.KafkaWriterConfigs {
		// create client
		c, err := NewKafkaClient(k)
		if err != nil {
			return err
		}
		p.KafkaProducers = append(p.KafkaProducers, c)
	}
	return nil
}
