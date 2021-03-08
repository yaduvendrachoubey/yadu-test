package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	"github.com/m3db/prometheus_client_golang/prometheus"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/ychoube/kafka-remote-write/config"
	"github.com/ychoube/kafka-remote-write/producer"
	"github.com/ychoube/kafka-remote-write/serializer"
	"github.com/ychoube/kafka-remote-write/server"
	"gopkg.in/alecthomas/kingpin.v2"
)

// FlagConfig is the
type FlagConfig struct {
	configFile       string
	testRun          bool
	listenAddr       string
	telemetryPath    string
	maxConnectionAge time.Duration
}

func main() {

	os.Setenv("PORT", "9000")

	var cfg FlagConfig
	cfg = FlagConfig{}

	app := kingpin.New(filepath.Base(os.Args[0]), "Remote Read/Write adaptor for prometheus")
	app.Version("kafka adaptor")
	app.HelpFlag.Short('h')

	app.Flag("config.file", "Remote Read/Write configuration file path.").
		Default("adaptor.yaml").StringVar(&cfg.configFile)
	app.Flag("web.listen-address", "Address to listen on for web endpoints.").
		Default(":9201").StringVar(&cfg.listenAddr)
	app.Flag("web.telemetry-path", "Path under which to expose metrics.").
		Default("/metrics").StringVar(&cfg.telemetryPath)
	app.Flag("web.max-connection-age", "If set this limits the maximum lifetime of persistent HTTP connections.").
		Default("0s").DurationVar(&cfg.maxConnectionAge)
	app.Flag("test", "Enable test mode").BoolVar(&cfg.testRun)

	_, err := app.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "Error parsing commandline arguments"))
		app.Usage(os.Args[1:])
		os.Exit(2)
	}
	adaptorConf, err := config.LoadConfigFile(cfg.configFile)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(adaptorConf)

	serializerFormat, err := serializer.ParseSerializationFormat(adaptorConf.GlobalConfig.SerializationFormat)
	if err != nil {
		fmt.Println("Error")
	}

	// initialize producer clients
	p := &producer.Producers{
		ProducersConfig: adaptorConf.RemoteWriteConfigs,
	}
	err = p.ProducerClients()
	if err != nil {
		fmt.Println(err)
	}

	r := server.NewServer()

	r.GinEngine.Use(ginrus.Ginrus(logrus.StandardLogger(), time.RFC3339, true), gin.Recovery())

	r.GinEngine.GET("/metrics", gin.WrapH(prometheus.UninstrumentedHandler()))
	r.GinEngine.POST("/receive", server.ReceiveHandler(p, serializerFormat))
	r.Run()

}
