package serializer

import (
	"encoding/json"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	"github.com/ychoube/kafka-remote-write/producer"

	"github.com/linkedin/goavro"
)

// Serializer represents an abstract metrics serializer
type Serializer interface {
	Marshal(metric map[string]interface{}) ([]byte, error)
}

func ParseSerializationFormat(value string) (Serializer, error) {
	switch value {
	case "json":
		return NewJSONSerializer()
	case "avro-json":
		return NewAvroJSONSerializer("schemas/metric.avsc")
	default:
		logrus.WithField("serialization-format-value", value).Warningln("invalid serialization format, using json")
		return NewJSONSerializer()
	}
}

// Serialize generates the JSON representation for a given Prometheus metric.
func Serialize(s Serializer, req *prompb.WriteRequest) (*producer.KafkaMessages, error) {
	result := &producer.KafkaMessages{}

	for _, ts := range req.Timeseries {
		labels := make(map[string]string, len(ts.Labels))

		for _, l := range ts.Labels {
			labels[string(model.LabelName(l.Name))] = string(model.LabelValue(l.Value))
		}

		for _, sample := range ts.Samples {
			epoch := time.Unix(sample.Timestamp/1000, 0).UTC()

			m := map[string]interface{}{
				"timestamp": epoch.Format(time.RFC3339),
				"value":     strconv.FormatFloat(sample.Value, 'f', -1, 64),
				"name":      string(labels["__name__"]),
				"labels":    labels,
			}

			data, err := s.Marshal(m)
			if err != nil {
				logrus.WithError(err).Errorln("couldn't marshal timeseries")
			}

			result.MessageList = append(result.MessageList, kafka.Message{
				Value: data,
			})

		}
	}

	return result, nil
}

// JSONSerializer represents a metrics serializer that writes JSON
type JSONSerializer struct {
}

func (s *JSONSerializer) Marshal(metric map[string]interface{}) ([]byte, error) {
	return json.Marshal(metric)
}

func NewJSONSerializer() (*JSONSerializer, error) {
	return &JSONSerializer{}, nil
}

// AvroJSONSerializer represents a metrics serializer that writes Avro-JSON
type AvroJSONSerializer struct {
	codec *goavro.Codec
}

func (s *AvroJSONSerializer) Marshal(metric map[string]interface{}) ([]byte, error) {

	return s.codec.BinaryFromNative(nil, metric)
}

// NewAvroJSONSerializer builds a new instance of the AvroJSONSerializer
func NewAvroJSONSerializer(schemaPath string) (*AvroJSONSerializer, error) {
	schema, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		logrus.WithError(err).Errorln("couldn't read avro schema")
		return nil, err
	}

	codec, err := goavro.NewCodec(string(schema))
	if err != nil {
		logrus.WithError(err).Errorln("couldn't create avro codec")
		return nil, err
	}

	return &AvroJSONSerializer{
		codec: codec,
	}, nil
}
