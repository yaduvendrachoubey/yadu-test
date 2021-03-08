// Copyright 2018 Telefónica
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/ychoube/kafka-remote-write/producer"
	"github.com/ychoube/kafka-remote-write/serializer"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/golang/snappy"

	"github.com/gogo/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/prompb"
)

var (
	httpRequestsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Count of all http requests",
		})
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
}

func processWriteRequest(req *prompb.WriteRequest, ser serializer.Serializer) ([][]byte, error) {
	return serializer.Serialize(ser, req)
}

func ReceiveHandler(producers *producer.Producers, ser serializer.Serializer) func(c *gin.Context) {
	return func(c *gin.Context) {

		httpRequestsTotal.Add(float64(1))

		compressed, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			logrus.WithError(err).Error("couldn't read body")
			return
		}

		reqBuf, err := snappy.Decode(nil, compressed)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			logrus.WithError(err).Error("couldn't decompress body")
			return
		}

		var req prompb.WriteRequest
		if err := proto.Unmarshal(reqBuf, &req); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			logrus.WithError(err).Error("couldn't unmarshal body")
			return
		}

		msg, err := processWriteRequest(&req, ser)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			logrus.WithError(err).Error("couldn't process write request")
			return
		}

		for _, p := range producers.KafkaProducers {
			for _, e := range msg {
				p.KafkaClient.Produce(&kafka.Message{
					TopicPartition: kafka.TopicPartition{Topic: &p.KafkaClientConfig.TopicName, Partition: kafka.PartitionAny},
					Value:          e,
				}, nil)
			}
		}

	}
}
