package support

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/snipa22/monerocnutils/serialization"
	"net"
)

type Pool struct {
	Identifier           string              // Unique identifier for the pool
	Hostname             string              // Hostname for the pool
	Port                 int32               // Port to connect to
	SSL                  bool                // Does this port use SSL?
	Share                int8                // Sub section of 100 to route traffic to.
	Username             string              // Username to connect to the pool with
	Password             string              // Password to connect to the pool with
	KeepAlive            bool                // Does this pool use keep alives?
	Primary              bool                // Is this the primary pool?
	Dev                  bool                // Is the the developer pool?
	PastBlockTemplates   []BlockTemplate     // Previous block templates
	CurrentBlockTemplate serialization.Block // Current block template
	Active               bool                // Is this pool active?
	SendId               uint64              // Number of messages sent by the pool
	SendLog              map[uint64]string   // Logs of all sent messages to this pool instance
	PoolJobs             map[uint64]MinerJob // Job for the miner
	AllowSelfSignedSSL   bool                // To allow a self-signed SSL or not
	Socket               net.Conn            // Active connection to the server
	XnpEnabled           bool                // Does this pool support the XNP extensions?
}

func (p *Pool) Connect() {
	var host = fmt.Sprintf("%v:%v", p.Hostname, p.Port)
	var err error
	if p.SSL {
		var conf = &tls.Config{
			InsecureSkipVerify: p.AllowSelfSignedSSL,
		}
		p.Socket, err = tls.Dial("tcp", host, conf)
	} else {
		p.Socket, err = net.Dial("tcp", host)
	}
	if err != nil {
		p.Active = false
	}
}

func (p *Pool) SendMessage(method string, params map[string]string) error {
	p.SendId += 1
	var rawSend map[string]string = map[string]string{
		"method": method,
		"id":     fmt.Sprintf("%v", p.SendId),
	}
	params["id"] = fmt.Sprintf("%v", p.SendId)
	jsonData, err := json.Marshal(params)
	if err != nil {
		return err
	}
	rawSend["params"] = string(jsonData)
	jsonData, err = json.Marshal(rawSend)
	if err != nil {
		return err
	}
	sendString := fmt.Sprintf("%v\n", string(jsonData))
	_, err = p.Socket.Write([]byte(sendString))
	if err != nil {
		return err
	}
	return nil
}

func (p *Pool) Login() {
	err := p.SendMessage("login", map[string]string{
		"login":    p.Username,
		"password": p.Password,
		"agent":    "xmr-node-proxy/0.0.3/compat/moneroGoProxy",
	})
	if err != nil {
		return
	}
	p.Active = true
}

func (p *Pool) Heartbeat() {
	if p.KeepAlive {
		_ = p.SendMessage("heartbeat", map[string]string{})
	}
}

func (p *Pool) SendShare(w Miner) {}
