package main

import (
	"encoding/hex"
	"encoding/json"
	"github.com/snipa22/moneroGoProxy/support"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"time"
)

// This project exists to provide a high speed proxy for the Monero Cryptocurrency.
// This package supersedes the xmr-node-proxy package

// The technical implementation of this design is slightly different than the original xmr-node-proxy
// This file launches 3 sets of Goroutines, one to connect to each pool and maintain the connection there
// The second set is used to handle inbound miner connections.
// The third set manages the actual connections from the miners, and the main listeners

// This file also contains the main HTTP API

func main() {
	var mainPool support.Pool
	var miners []support.Miner
	var ports []support.Port
	l := make(chan net.Conn, 32) // Inbound Listeners, convert to miners
	mr := make(chan string, 32)  // Miner receive channel to announce things are done
	for {
		select {
		case c := <-l:
			// Setup new miner.
			var p support.Port
			for p, _ = range ports {
				if p.Port == uint32(c.LocalAddr().(*net.TCPAddr).Port) {
					break
				}
			}
			token := make([]byte, 21)
			rand.Read(token)
			m := support.Miner{
				Socket:        c,
				Pool:          mainPool,
				ConnectTime:   time.Time{},
				Port:          p,
				Difficulty:    p.StartingDiff,
				FixedDiff:     p.FixedDiff,
				LastContact:   time.Time{},
				LastShareTime: time.Time{},
				CachedJob:     support.MinerJob{},
				Alive:         false,
				ID:            hex.EncodeToString(token),
			}
			miners = append(miners, m)
			go m.MinerEntry(mr)
		}
	}
}

type ProxyConfig struct {
	Pools []support.Pool // Pool configurations
	Ports []support.Port // Listening port configurations
}

func loadConfig() ProxyConfig {
	jsonFile, err := os.Open("config.json")

	if err != nil {
		log.Fatal(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var config ProxyConfig

	_ = json.Unmarshal(byteValue, &config)
}
