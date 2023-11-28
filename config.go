package main

import (
	"errors"
	"flag"
	"os"

	logrus "github.com/sirupsen/logrus"
)

type Config struct {
	PubSubTopic               string
	ContentTopic              string
	QueriesPerSecond          uint64
	NumMinutesQuery           uint64 // we perform a query to get msgs from now until -NumMinutesQuery back
	BootstrapNode             string
	PublishInvalid            bool
	LogSentMessages           bool
	PeerStoreSQLiteAddrList   string // comma-separated list of multiaddresses of Store nodes with SQLite
	PeerStorePostgresAddrList string // comma-separated list of multiaddresses of Store nodes with PostgreSQL
	NumUsers                  uint64
}

// By default the release is a custom build. CI takes care of upgrading it with
// go build -v -ldflags="-X 'github.com/xxx/yyy/config.ReleaseVersion=x.y.z'"
var ReleaseVersion = "custom-build"

func NewCliConfig() (*Config, error) {
	var version = flag.String("version", "", "Prints the release version and exits")
	var pubSubTopic = flag.String("pubsub-topic", "/waku/2/default-waku/proto", "PubSub topic to make Store requests to")
	var contentTopic = flag.String("content-topic", "my-ctopic", "Content topic to make Store requests to")
	var queriesPerSecond = flag.Uint64("queries-per-second", 0, "Number of queries to be made per second")
	var bootstrapNode = flag.String("bootstrap-node", "", "Bootstrap node to connect to")
	var logSentMessages = flag.Bool("log-sent-messages", false, "Logs the messages that are sent. default: false")
	var peerStoreSQLiteAddrList = flag.String("peer-store-sqlite-addr-list", "", "comma-separated multiaddress of peer with Store mounted and using SQLite as archiver")
	var peerStorePostgresAddrList = flag.String("peer-store-postgres-addr-list", "", "comma-separated multiaddress of peer with Store mounted and using Postgres as archiver")
	var numMinutes = flag.Uint64("num-minutes-query", 5, "Defines the message range (storedAt) in the Store query")
	var numUsers = flag.Uint64("num-concurrent-users", 1, "Defines the number of concurrent users")

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

	if *peerStoreSQLiteAddrList == "" && *peerStorePostgresAddrList == "" {
		return nil, errors.New("either --peer-store-sqlite-addr-list or --peer-store-postgres-addr-list is required")
	}

	conf := &Config{
		PubSubTopic:               *pubSubTopic,
		ContentTopic:              *contentTopic,
		QueriesPerSecond:          *queriesPerSecond,
		NumMinutesQuery:           *numMinutes,
		BootstrapNode:             *bootstrapNode,
		LogSentMessages:           *logSentMessages,
		PeerStoreSQLiteAddrList:   *peerStoreSQLiteAddrList,
		PeerStorePostgresAddrList: *peerStorePostgresAddrList,
		NumUsers:                  *numUsers,
	}
	logConfig(conf)
	return conf, nil
}

func logConfig(cfg *Config) {
	logrus.WithFields(logrus.Fields{
		"PubSubTopic":               cfg.PubSubTopic,
		"ContentTopic":              cfg.ContentTopic,
		"QueriesPerSecond":          cfg.QueriesPerSecond,
		"NumMinutesQuery":           cfg.NumMinutesQuery,
		"BootstrapNode":             cfg.BootstrapNode,
		"LogSentMessages":           cfg.LogSentMessages,
		"PeerStoreSQLiteAddrList":   cfg.PeerStoreSQLiteAddrList,
		"PeerStorePostgresAddrList": cfg.PeerStorePostgresAddrList,
		"NumUsers":                  cfg.NumUsers,
	}).Info("Cli Config:")
}
