package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
	"sync"

	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	log "github.com/sirupsen/logrus"
	"github.com/waku-org/go-waku/waku/v2/node"
	"github.com/waku-org/go-waku/waku/v2/protocol/store"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap"
)

var storeReqSent = 0

func newConnManager(lo int, hi int, opts ...connmgr.Option) *connmgr.BasicConnMgr {
	mgr, err := connmgr.NewConnManager(lo, hi, opts...)
	if err != nil {
		panic("could not create ConnManager: " + err.Error())
	}
	return mgr
}

func main() {
	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.FullTimestamp = true
	log.SetFormatter(customFormatter)

	cfg, err := NewCliConfig()
	if err != nil {
		log.Fatal("error parsing config: ", err)
	}

	hostAddr, _ := net.ResolveTCPAddr("tcp", "0.0.0.0:0")
	key, err := randomHex(32)
	if err != nil {
		log.Error("Could not generate random key: ", err)
		return
	}
	prvKey, err := crypto.HexToECDSA(key)
	if err != nil {
		log.Error("Could not convert hex into ecdsa key: ", err)
		return
	}

	ctx := context.Background()

	customNodes := []*enode.Node{
		enode.MustParse(cfg.BootstrapNode),
	}

	wakuNode, err := node.New(
		node.WithPrivateKey(prvKey),
		node.WithHostAddress(hostAddr),
		node.WithNTP(),
		node.WithWakuRelay(),
		node.WithDiscoveryV5(8000, customNodes, true),
		node.WithDiscoverParams(30),
		node.WithLogLevel(zapcore.Level(zapcore.ErrorLevel)),
	)
	if err != nil {
		log.Error("Error creating wakunode: ", err)
		return
	}

	if err := wakuNode.Start(ctx); err != nil {
		log.Error("Error starting wakunode: ", err)
		return
	}

	err = wakuNode.DiscV5().Start(ctx)
	if err != nil {
		log.Fatal("Error starting discovery: ", err)
	}

	// Connecting to nodes
	// ================================================================

	log.Info("Connecting to nodes...")

	peers := []string{
		cfg.PeerStoreSQLiteAddr,
		cfg.PeerStorePostgresAddr,
	}

	connectToNodes(ctx, wakuNode, peers)

	time.Sleep(2 * time.Second) // Required so Identify protocol is executed

	go logPeriodicInfo(wakuNode)
	go runEvery(wakuNode, peers, int(cfg.QueriesPerSecond))

	// Wait for a SIGINT or SIGTERM signal
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	fmt.Println("\n\n\nReceived signal, shutting down...")

	// shut the node down
	wakuNode.Stop()
}

func logPeriodicInfo(wakuNode *node.WakuNode) {
	for {
		fmt.Println("Connected peers", wakuNode.PeerCount(), " msg sent: ", storeReqSent)
		time.Sleep(10 * time.Second)
	}
}

func randomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func runEvery(wakuNode *node.WakuNode, peers []string, queriesPerSec int) {
	tickInSeconds := int64(1)
	ticker := time.NewTicker(time.Duration(tickInSeconds) * time.Second)
	quit := make(chan struct{})
	for {
		select {
		case <-ticker.C:
			start := time.Now().UnixNano() / int64(time.Millisecond)

			// Not very fancy. requests are not evenly distributed over time
			// eg 5 msg per second are send very quickly and the remaning time to complete
			// the second is idle.
			for i := 0; i < queriesPerSec; i++ {
				performStoreQueries(wakuNode, peers)
				storeReqSent++
			}
			end := time.Now().UnixNano() / int64(time.Millisecond)

			diff := end - start
			fmt.Println("Duration(ms):", diff)
			// do something if request rate is greater than what can be handled
			if diff > tickInSeconds*1000 {
				fmt.Println("Warning: took more than 1 second")
			}
		case <-quit:
			ticker.Stop()
			return
		}
	}
}

func connectToNodes(ctx context.Context,
					node *node.WakuNode,
					nodeList []string) {
	wg := sync.WaitGroup{}
	for _, addr := range nodeList {
		wg.Add(1)
		go func(addr string) {
			wg.Done()
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			err := node.DialPeer(ctx, addr)
			if err != nil {
				log.Error("could not connect to peer", zap.String("addr", addr), zap.Error(err))
			}
		}(addr)
	}
	wg.Wait()
}

func queryNode(ctx context.Context,
			   node *node.WakuNode,
			   addr string,
			   pubsubTopic string,
			   contentTopic string,
			   startTime time.Time,
			   endTime time.Time) (int, error) {

	p, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		return -1, err
	}

	info, err := peer.AddrInfoFromP2pAddr(p)
	if err != nil {
		return -1, err
	}

	cnt := 0
	cursorIterations := 0

	requestId := uuid.New()

	result, err := node.Store().Query(ctx,
									  store.Query{
											Topic:         pubsubTopic,
											ContentTopics: []string{contentTopic},
											StartTime:     startTime.UnixNano(),
											EndTime:       endTime.UnixNano(),
										},
										store.WithPeer(info.ID), store.WithPaging(false, 100),
										store.WithRequestId([]byte(requestId.String())),
									)
	if err != nil {
		return -1, err
	}

	for {
		hasNext, err := result.Next(ctx)
		if err != nil {
			return -1, err
		}

		if !hasNext { // No more messages available
			break
		}
		cursorIterations += 1

		cnt += len(result.GetMessages())
	}

	log.Info(fmt.Sprintf("%d messages found in %s (Used cursor %d times)\n",
			 cnt, info.ID, cursorIterations))

	return cnt, nil
}

// This func performs a query to both the SQLite and Postgres Store peers.
// Notice that these peers should be passed as a parameter to the app.
func performStoreQueries(
	wakuNode *node.WakuNode,
	peers []string) {

	pubsubTopic := "/waku/2/default-waku/proto"
	endTime := wakuNode.Timesource().Now()
	startTime := endTime.Add(-time.Minute * 5)
	contentTopic := "my-ctopic"

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second) // Test shouldnt take more than 60s
	defer cancel()

	// Store
	// ================================================================

	timeSpentMap := make(map[string]time.Duration)
	numUsers := int64(10)

	wg := sync.WaitGroup{}
	for _, addr := range peers {
		for userIndex := 0; userIndex < int(numUsers); userIndex++ {
			wg.Add(1)
			go func(addr string) {
				defer wg.Done()
				fmt.Println("Querying node", addr)
				start := time.Now()
				cnt, _ := queryNode(ctx, wakuNode, addr, pubsubTopic, contentTopic, startTime, endTime)
				timeSpent := time.Since(start)
				fmt.Printf("\n%s took %v. Obtained: %d rows", addr, timeSpent, cnt)
				timeSpentMap[addr] += timeSpent
			}(addr)
		}
	}

	wg.Wait()

	timeSpentNanos := timeSpentMap[peers[0]].Nanoseconds() / numUsers
	fmt.Println("\n\nAverage time spent: ", peers[0], time.Duration(timeSpentNanos))

	timeSpentNanos = timeSpentMap[peers[1]].Nanoseconds() / numUsers
	fmt.Println("\n\nAverage time spent:", peers[1], time.Duration(timeSpentNanos))
	fmt.Println("")
}
