package main

import (
	"errors"
	"flag"
	"os"

	logrus "github.com/sirupsen/logrus"
)

type Config struct {
	PubSubTopic     string
	ContentTopic    string
	QueriesPerSecond    uint64
	NumMinutesQuery    uint64 // we perform a query to get msgs from now until -NumMinutesQuery back
	BootstrapNode   string
	PublishInvalid  bool
	LogSentMessages bool
	PeerStoreSQLiteAddr string
	PeerStorePostgresAddr string
}

// By default the release is a custom build. CI takes care of upgrading it with
// go build -v -ldflags="-X 'github.com/xxx/yyy/config.ReleaseVersion=x.y.z'"
var ReleaseVersion = "custom-build"

func NewCliConfig() (*Config, error) {
	var version = flag.String("version", "", "Prints the release version and exits")
	var pubSubTopic = flag.String("pubsub-topic", "", "PubSub topic to make Store requests to")
	var contentTopic = flag.String("content-topic", "", "Content topic to make Store requests to")
	var queriesPerSecond = flag.Uint64("queries-per-second", 0, "Number of queries to be made per second")
	var bootstrapNode = flag.String("bootstrap-node", "", "Bootstrap node to connect to")
	var logSentMessages = flag.Bool("log-sent-messages", false, "Logs the messages that are sent. default: false")
	var peerStoreSQLiteAddr = flag.String("peer-store-sqlite-addr", "", "Multiaddress of peer with Store mounted and using SQLite as archiver")
	var peerStorePostgresAddr = flag.String("peer-store-postgres-addr", "", "Multiaddress of peer with Store mounted and using Postgres as archiver")
	var numMinutes = flag.Uint64("num-minutes-query", 5, "Defines the message range (storedAt) in the Store query")

	flag.Parse()

	if *version != "" {
		logrus.Info("Version: ", ReleaseVersion)
		os.Exit(0)
	}

	// Some simple validation
	if *pubSubTopic == "" {
		return nil, errors.New("--pubsub-topic is required")
	}

	if *contentTopic == "" {
		return nil, errors.New("--content-topic is required")
	}

	if *queriesPerSecond == 0 {
		return nil, errors.New("--queries-per-second is required")
	}

	if *bootstrapNode == "" {
		return nil, errors.New("--bootstrap-node is required")
	}

	if *peerStoreSQLiteAddr == "" {
		return nil, errors.New("--peer-store-sqlite-addr is required")
	}

	if *peerStorePostgresAddr == "" {
		return nil, errors.New("--peer-store-postgres-addr is required")
	}

	conf := &Config{
		PubSubTopic:     *pubSubTopic,
		ContentTopic:    *contentTopic,
		QueriesPerSecond:    *queriesPerSecond,
		NumMinutesQuery: *numMinutes,
		BootstrapNode:   *bootstrapNode,
		LogSentMessages: *logSentMessages,
		PeerStoreSQLiteAddr: *peerStoreSQLiteAddr,
		PeerStorePostgresAddr: *peerStorePostgresAddr,
	}
	logConfig(conf)
	return conf, nil
}

func logConfig(cfg *Config) {
	logrus.WithFields(logrus.Fields{
		"PubSubTopic":     cfg.PubSubTopic,
		"ContentTopic":    cfg.ContentTopic,
		"QueriesPerSecond":    cfg.QueriesPerSecond,
		"NumMinutesQuery": cfg.NumMinutesQuery,
		"BootstrapNode":   cfg.BootstrapNode,
		"LogSentMessages": cfg.LogSentMessages,
		"PeerStoreSQLiteAddr": cfg.PeerStoreSQLiteAddr,
		"PeerStorePostgresAddr": cfg.PeerStorePostgresAddr,
	}).Info("Cli Config:")
}
