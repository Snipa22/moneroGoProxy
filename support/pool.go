package support

import (
	"container/ring"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/snipa22/monerocnutils/serialization"
	"net"
)

type Pool struct {
	identifier           string              // Unique identifier for the pool
	hostname             string              // Hostname for the pool
	port                 int32               // Port to connect to
	ssl                  bool                // Does this port use SSL?
	share                int8                // Sub section of 100 to route traffic to.
	username             string              // Username to connect to the pool with
	password             string              // Password to connect to the pool with
	keepAlive            bool                // Does this pool use keep alives?
	primary              bool                // Is this the primary pool?
	dev                  bool                // Is the the developer pool?
	pastBlockTemplates   ring.Ring           // Previous block templates
	currentBlockTemplate serialization.Block // Current block template
	active               bool                // Is this pool active?
	sendId               uint64              // Number of messages sent by the pool
	sendLog              map[uint64]string   // Logs of all sent messages to this pool instance
	poolJobs             map[uint64]Job      // Job by miner
	allowSelfSignedSSL   bool                // To allow a self-signed SSL or not
	socket               net.Conn            // Active connection to the server
	xnpEnabled           bool                // Does this pool support the XNP extensions?
}

func (p *Pool) Connect() {
	var host = fmt.Sprintf("%v:%v", p.hostname, p.port)
	var err error
	if p.ssl {
		var conf = &tls.Config{
			InsecureSkipVerify: p.allowSelfSignedSSL,
		}
		p.socket, err = tls.Dial("tcp", host, conf)
	} else {
		p.socket, err = net.Dial("tcp", host)
	}
	if err != nil {
		p.active = false
	}
}

func (p *Pool) SendMessage(method string, params map[string]string) error {
	p.sendId += 1
	var rawSend map[string]string = map[string]string{
		"method": method,
		"id":     fmt.Sprintf("%v", p.sendId),
	}
	params["id"] = fmt.Sprintf("%v", p.sendId)
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
	_, err = p.socket.Write([]byte(sendString))
	if err != nil {
		return err
	}
}

func (p *Pool) Login() {
	err := p.SendMessage("login", map[string]string{
		"login":    p.username,
		"password": p.password,
		"agent":    "xmr-node-proxy/0.0.3/compat/moneroGoProxy",
	})
	if err != nil {
		return
	}
	p.active = true
}

func (p *Pool) Heartbeat() {
	if p.keepAlive {
		_ = p.SendMessage("heartbeat", map[string]string{})
	}
}

func (p *Pool) SendShare(w Miner) {}
