package smegdns

import (
	"encoding/json"
	"fmt"
	"net"
	"net/rpc"

	"github.com/miekg/dns"
	"github.com/tim-beatham/wgmesh/pkg/ipc"
	"github.com/tim-beatham/wgmesh/pkg/lib"
	logging "github.com/tim-beatham/wgmesh/pkg/log"
	"github.com/tim-beatham/wgmesh/pkg/query"
)

const SockAddr = "/tmp/wgmesh_ipc.sock"

const MeshRegularExpression = `(?P<meshId>.+)\.(?P<alias>.+)\.smeg\.`

type DNSHandler struct {
	client *rpc.Client
	server *dns.Server
}

// queryMesh: queries the mesh network for the given meshId and node
// with alias
func (d *DNSHandler) queryMesh(meshId, alias string) net.IP {
	var reply string

	err := d.client.Call("IpcHandler.Query", &ipc.QueryMesh{
		MeshId: meshId,
		Query:  "[?alias == 'tim'] | [0]",
	}, &reply)

	if err != nil {
		return nil
	}

	var node *query.QueryNode

	err = json.Unmarshal([]byte(reply), &node)

	if err != nil || node == nil {
		return nil
	}

	ip, _, _ := net.ParseCIDR(node.WgHost)
	return ip
}

func (d *DNSHandler) handleQuery(m *dns.Msg) {
	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeAAAA:
			logging.Log.WriteInfof("Query for %s", q.Name)

			groups := lib.MatchCaptureGroup(MeshRegularExpression, q.Name)

			if len(groups) == 0 {
				continue
			}

			ip := d.queryMesh(groups["meshId"], groups["alias"])

			rr, err := dns.NewRR(fmt.Sprintf("%s AAAA %s", q.Name, ip))

			if err != nil {
				logging.Log.WriteErrorf(err.Error())
			}

			if err == nil {
				m.Answer = append(m.Answer, rr)
			}
		}
	}
}

func (h *DNSHandler) handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.Authoritative = true

	switch r.Opcode {
	case dns.OpcodeQuery:
		h.handleQuery(msg)
	}

	w.WriteMsg(msg)
}

func (h *DNSHandler) Listen() error {
	return h.server.ListenAndServe()
}

func (h *DNSHandler) Close() error {
	return h.server.Shutdown()
}

func NewDns(udpPort int) (*DNSHandler, error) {
	client, err := rpc.DialHTTP("unix", SockAddr)

	if err != nil {
		return nil, err
	}

	dnsHander := DNSHandler{
		client: client,
	}

	dns.HandleFunc("smeg.", dnsHander.handleDnsRequest)

	dnsHander.server = &dns.Server{Addr: fmt.Sprintf(":%d", udpPort), Net: "udp"}
	return &dnsHander, nil
}
