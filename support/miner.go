package support

import (
	"container/ring"
	"net"
	"time"
)

type Miner struct {
	login         string    // Not used
	password      string    // Not used
	agent         string    // Not used
	ip            string    // Not used
	socket        net.Conn  // Communication with the miner
	pool          Pool      // What pool is the miner currently active on?
	connectTime   time.Time // When did the miner connect
	difficulty    uint64    // Target?
	fixedDiff     bool      // Shifting algo?  Y/N
	incremented   bool      // Hrm.
	shares        uint64    // How many shares completed?
	blocks        uint64    // How many blocks found?
	hashes        uint64    // Whats the total hash count?
	lastContact   time.Time // Last time the miner was in contact
	lastShareTime time.Time // When was the miner's last share?
	validJobs     ring.Ring
}
