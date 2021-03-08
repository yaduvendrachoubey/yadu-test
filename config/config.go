package config

import (
	"io/ioutil"
	"time"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

// GlobalConfig configures global settings for adaptor
type GlobalConfig struct {
	WriteTimeout        time.Duration `yaml:"write_timeout"`
	ReadTimeout         time.Duration `yaml:"read_timeout"`
	SerializationFormat string        `yaml:"serialization_format"`
}

// RemoteWriteConfig configures remote writer for adaptor
type RemoteWriteConfig struct {
	KafkaWriterConfigs        []*KafkaWriterConfig        `yaml:"kafka_writers", omitempty`
	GoogleCloudStorageConfigs []*GoogleCloudStorageConfig `yaml:"gcs_writers",omitempty`
}

// ProducerConfig configures producer for kafka writer
type ProducerConfig struct {
	MaxBatchSize int    `yaml:"max_batch_size",omitempty`
	MessageKey   string `yaml:"message_key",omitempty`
}

// KafkaWriterConfig configures writers for kafka
type KafkaWriterConfig struct {
	JobName         string         `yaml:"job_name,omitempty"`
	KafkaHosts      []string       `yaml:"kafka_hosts,omitempty"`
	TopicName       string         `yaml:"topic_name,omitempty"`
	ProducerConfigs ProducerConfig `yaml:"producer_config"`
}

type GoogleCloudStorageConfig struct {
	Dest string `yaml:"gcs,omitempty"`
}

type Config struct {
	GlobalConfig       GlobalConfig      `yaml:"global"`
	RemoteWriteConfigs RemoteWriteConfig `yaml:"remote_write,omitempty"`
	//RemoteReadConfigs  []*RemoteReadConfig  `yaml:"remote_read,omitempty"`
}

// LoadConfigContent loads the YAML config
func LoadConfigContent(s []byte) (*Config, error) {
	cfg := &Config{}

	err := yaml.UnmarshalStrict([]byte(s), cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil

}

func LoadConfigFile(filename string) (*Config, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	cfg, err := LoadConfigContent(content)
	if err != nil {
		return nil, errors.Wrapf(err, "Config file error %s", filename)
	}
	return cfg, nil

}
