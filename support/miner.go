package support

import (
	"encoding/json"
	"fmt"
	"math"
	"net"
	"time"
)

var MinDiff uint32 = 1000
var MaxDiff uint32 = 100000

type MinerRPCError struct {
	Code    int8   `json:"code"`    // Error code number, usually -1
	Message string `json:"message"` // Error message to send the miner
}

type MinerRPCResponse struct {
	ID      uint32             `json:"id"`                 // ID for the message
	JsonRPC string             `json:"jsonrpc"`            // The RPC version in use
	Error   *MinerRPCError     `json:"error,omitifempty"`  // Send error if we've got it.
	Result  *map[string]string `json:"result,omitifempty"` // String of the response to send as the body, usually a json object of some sort
}

type Miner struct {
	Login         string    // Not used
	Password      string    // Not used
	Agent         string    // Not used
	IP            string    // Not used
	Socket        net.Conn  // Communication with the miner
	Pool          Pool      // What pool is the miner currently active on?
	ConnectTime   time.Time // When did the miner connect
	Difficulty    uint32    // Target?
	FixedDiff     bool      // Shifting algo?  Y/N
	Incremented   bool      // Hrm.
	Shares        uint64    // How many shares completed?
	Blocks        uint64    // How many blocks found?
	Hashes        uint64    // Whats the total hash count?
	LastContact   time.Time // Last time the miner was in contact
	LastShareTime time.Time // When was the miner's last share?
	NewDiff       uint32    // Set a new diff on the next pass?
	CachedJob     MinerJob  // Current cached job
	RPCID         uint32    // Last known RPC ID from the miner
	Alive         bool      // Is the miner alive?
}

func (m *Miner) UpdateDifficulty(t uint64) {
	if m.Hashes > 0 && !m.FixedDiff {
		m.SetNewDiff(uint32(m.Hashes / uint64(math.Floor(time.Now().Sub(m.ConnectTime).Seconds())) * t))
	}
}

func (m *Miner) SetNewDiff(d uint32) {
	if d < MinDiff {
		d = MinDiff
	}
	if d > MaxDiff {
		d = MaxDiff
	}
	if m.Difficulty == d {
		return
	}
	m.NewDiff = d
}

// Take Error, Miner Respone ID
func (m *Miner) SendMessage(e error, r map[string]string) {
	mr := MinerRPCResponse{
		ID:      m.RPCID,
		JsonRPC: "",
		Error:   nil,
		Result:  nil,
	}
	if e != nil {
		mr.Error = &MinerRPCError{
			Code:    -1,
			Message: e.Error(),
		}
		rs, err := json.Marshal(mr)
		if err != nil {
			fmt.Println(err)
		}
		_, err = m.Socket.Write([]byte(fmt.Sprintf("%s\n", rs)))
		if err != nil {
			m.Alive = false
		}
		return
	}
	mr.Result = &r
	rs, err := json.Marshal(mr)
	if err != nil {
		fmt.Println(err)
	}
	_, err = m.Socket.Write([]byte(fmt.Sprintf("%s\n", rs)))
	if err != nil {
		m.Alive = false
	}
	return
}

type MinerJob struct {
	ID            string        // Job Identifier
	BlockTemplate BlockTemplate // Block data object that this job belongs to
	HashBlob      string        // Thing to go hash
	Target        string        // Difficulty target in hex
	JobNonce      uint32        // Nonce for the job
	Difficulty    uint32        // Miner Difficulty
	Submissions   []string      // List of all nonces submitted
}
