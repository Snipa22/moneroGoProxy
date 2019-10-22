package support

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

var MinDiff uint32 = 1000
var MaxDiff uint32 = 100000

var (
	ErrorNotAuthenticated = errors.New("not authenticated")
)

type MinerRPCError struct {
	Code    int8   `json:"code"`    // Error code number, usually -1
	Message string `json:"message"` // Error message to send the miner
}

type MinerRPCResponse struct {
	ID      uint32                  `json:"id"`                 // ID for the message
	JsonRPC string                  `json:"jsonrpc"`            // The RPC version in use
	Error   *MinerRPCError          `json:"error,omitifempty"`  // Send error if we've got it.
	Result  *map[string]interface{} `json:"result,omitifempty"` // String of the response to send as the body, usually a json object of some sort
}

type MinerRPCReceive struct {
	ID     uint32            `json:"id"`
	Method string            `json:"method"`
	Params map[string]string `json:"params"`
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
	ID            string    // Miner Identifier
}

func MinerEntry(c net.Conn, p Pool) {
	token := make([]byte, 21)
	rand.Read(token)

	m := Miner{
		Login:         "",
		Password:      "",
		Agent:         "",
		IP:            "",
		Socket:        c,
		Pool:          p,
		ConnectTime:   time.Time{},
		Difficulty:    0,
		FixedDiff:     false,
		Incremented:   false,
		Shares:        0,
		Blocks:        0,
		Hashes:        0,
		LastContact:   time.Time{},
		LastShareTime: time.Time{},
		NewDiff:       0,
		CachedJob:     MinerJob{},
		RPCID:         0,
		Alive:         false,
		ID:            hex.EncodeToString(token),
	}

	buf := bufio.NewReader(m.Socket)

	for {
		data, _, err := buf.ReadLine()
		if err != nil {
			return
		}
		m.HandleMessage(data)
		if !m.Alive {
			break
		}
	}
}

func (m *Miner) HandleMessage(msg []byte) {
	r := MinerRPCReceive{}
	err := json.Unmarshal(msg, r)
	if err != nil {
		fmt.Printf("Error in decoding message: %v from miner at %v (%v/%v)", msg, m.IP, m.Login, m.Password)
		if !m.Alive {
			_ = m.Socket.Close()
		}
		return
	}

	m.LastContact = time.Now()

	if r.ID > 0 {
		m.RPCID = r.ID
	}

	switch r.Method {
	case "login":
		// Handle the login case
		// The data files should be login, pass, agent
		m.Agent = r.Params["agent"]
		m.Login = r.Params["login"]
		m.Password = r.Params["pass"]

		pt := strings.Split(m.Login, "+")
		if len(pt) == 2 {
			m.FixedDiff = true
			fixedDiff, _ := strconv.Atoi(pt[1])
			m.Difficulty = uint32(fixedDiff)
		} else if len(pt) > 2 {
			m.SendMessage(errors.New("too many options in the login field"), map[string]interface{}{})
			return
		}
		m.SendMessage(nil, map[string]interface{}{
			"id":     m.ID,
			"status": "OK",
			"job":    m.Pool.CurrentBlockTemplate.GetJob(m, true).GetMinerSend(*m),
		})
		m.Alive = true
		break
	case "getjob":
		if !m.Alive {
			m.SendMessage(ErrorNotAuthenticated, map[string]interface{}{})
			return
		}
		// Hand them a new job damnit!
		m.SendMessage(nil, m.Pool.CurrentBlockTemplate.GetJob(m, true).GetMinerSend(*m))
		break
	case "submit":
		// Hash, hash, hash
		break
	case "keepalived":
		if !m.Alive {
			m.SendMessage(ErrorNotAuthenticated, map[string]interface{}{})
			return
		}
		// More keepalive.  Sigh
		m.SendMessage(nil, map[string]interface{}{
			"status": "KEEPALIVED",
		})
		break
	}
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
func (m *Miner) SendMessage(e error, r map[string]interface{}) {
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

func (mj MinerJob) GetMinerSend(m Miner) map[string]interface{} {
	return map[string]interface{}{
		"blob":   mj.HashBlob,
		"job_id": mj.ID,
		"target": mj.Target,
		"id":     m.ID,
		"height": mj.BlockTemplate.Height,
	}
}
