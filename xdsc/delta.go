package xdsc

import (
	"fmt"
	"log"
	"math"
	"sort"
	"strings"

	"github.com/costinm/istiod/xdsc/sets"
	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	v3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"google.golang.org/grpc"
)

type ResourceKey struct {
	Name    string
	TypeUrl string
}

func (k ResourceKey) String() string {
	if k == (ResourceKey{}) {
		return "<wildcard>"
	}
	return strings.TrimPrefix(k.TypeUrl, "type.googleapis.com/envoy.config.") + "/" + k.Name
}

type ResourceNode struct {
	Key ResourceKey

	Parents  map[*ResourceNode]struct{}
	Children map[*ResourceNode]struct{}
}

func DialDelta(url string, opts *Config) (*ADSC, error) {
	ListenerNode := &ResourceNode{
		Key:      ResourceKey{TypeUrl: v3.ListenerType},
		Parents:  map[*ResourceNode]struct{}{},
		Children: map[*ResourceNode]struct{}{},
	}

	ClusterNode := &ResourceNode{
		Key:      ResourceKey{TypeUrl: v3.ClusterType},
		Parents:  map[*ResourceNode]struct{}{},
		Children: map[*ResourceNode]struct{}{},
	}

	conn, err := grpc.DialContext(opts.Context, url, opts.GrpcOpts...)
	if err != nil {
		return nil, fmt.Errorf("dial context: %v", err)
	}

	xds := discovery.NewAggregatedDiscoveryServiceClient(conn)
	xdsClient, err := xds.DeltaAggregatedResources(opts.Context, grpc.MaxCallRecvMsgSize(math.MaxInt32))
	if err != nil {
		return nil, fmt.Errorf("stream: %v", err)
	}
	c := &ADSC{
		initialWatches: []string{v3.ClusterType, v3.ListenerType},
		conn:           conn,
		client:         xdsClient,
		resources:      map[string]sets.String{},
		tree: map[ResourceKey]*ResourceNode{
			ListenerNode.Key: ListenerNode,
			ClusterNode.Key:  ClusterNode,
		},
	}
	c.node = c.makeNode()
	go c.handleRecv()
	return c, nil
}

func (d *ADSC) handleDeltaRecv() {
	for {
		msg, err := d.client.Recv()
		if err != nil {
			log.Printf("Connection closed: %v", err)
			d.Close()
			return
		}

		requests := map[string][]string{}
		resources := sets.New[string]()

		d.mu.Lock()
		for _, resp := range msg.Resources {
			key := ResourceKey{
				Name:    resp.Name,
				TypeUrl: msg.TypeUrl,
			}

			if d.tree[key] == nil && isWildcardTypeURL(msg.TypeUrl) {
				d.tree[key] = &ResourceNode{
					Key:      key,
					Parents:  map[*ResourceNode]struct{}{},
					Children: map[*ResourceNode]struct{}{},
				}
				switch msg.TypeUrl {
				case v3.ListenerType:
					relate(d.tree[ResourceKey{TypeUrl: v3.ListenerType}], d.tree[key])
				case v3.ClusterType:
					relate(d.tree[ResourceKey{TypeUrl: v3.ClusterType}], d.tree[key])
				}

			} else if d.tree[key] == nil {
				log.Printf("Ignoring unmatched resource %s", key)
				continue
			}

			node := d.tree[key]
			resources = resources.Insert(resp.Name)
			referenced := extractReferencedKeys(resp)
			for _, rkey := range referenced {
				child, f := d.getNode(rkey)
				if !f {
					requests[rkey.TypeUrl] = append(requests[rkey.TypeUrl], rkey.Name)
				}
				relate(node, child)
			}
		}
		removals := map[string][]string{}
		for _, resp := range msg.RemovedResources {
			key := ResourceKey{
				Name:    resp,
				TypeUrl: msg.TypeUrl,
			}
			if d.tree[key] == nil {
				log.Printf("Ignoring removing unmatched resource %s, %v", key, d.dumpTree())
				continue
			}
			node := d.tree[key]
			d.deleteNode(node, removals)
		}

		//origLen := len(d.resources[msg.TypeUrl])
		newAdd := d.resources[msg.TypeUrl].Union(resources)
		//addedLen := len(newAdd) - origLen
		//removedLen := len(msg.RemovedResources)
		d.resources[msg.TypeUrl] = newAdd.Difference(sets.New(msg.RemovedResources...))
		d.mu.Unlock()
		//log.WithLabels("type", msg.TypeUrl, "added", addedLen, "removed", removedLen, "removed refs", len(removals)).Debugf("got message")
		//if dumpScope.DebugEnabled() {
		//	s, _ := (&jsonpb.Marshaler{}).MarshalToString(msg)
		//	dumpScope.Debug(s)
		//}
		//if dumpScope.InfoEnabled() {
		//	d.mu.Lock()
		//	dumpScope.Info("\n" + d.dumpTree())
		//	d.mu.Unlock()
		//}

		// TODO: Envoy does some smart "pausing" to allow the next push to come before we request
		for _, k := range keysOfMaps(requests, removals) {
			if len(requests[k]) == 0 && len(removals[k]) == 0 {
				continue
			}
			if err := d.dsend(&discovery.DeltaDiscoveryRequest{
				TypeUrl:                  k,
				ResourceNamesSubscribe:   requests[k],
				ResourceNamesUnsubscribe: removals[k],
			}, ReasonRequest); err != nil {
				log.Printf("error sending request: %v", err)
			}
		}

		if err := d.dsend(&discovery.DeltaDiscoveryRequest{
			TypeUrl:       msg.TypeUrl,
			ResponseNonce: msg.Nonce,
		}, ReasonAck); err != nil {
			log.Printf("error sending ACK: %v", err)
		}
	}
}

func keysOfMaps(ms ...map[string][]string) []string {
	res := []string{}
	for _, m := range ms {
		for k := range m {
			res = append(res, k)
		}
	}
	// TODO sort in XDS ordering?
	sort.Strings(res)
	return res
}

func extractReferencedKeys(resp *discovery.Resource) []ResourceKey {
	res := []ResourceKey{}
	switch resp.Resource.TypeUrl {
	case v3.ClusterType:
		o := &cluster.Cluster{}
		_ = resp.Resource.UnmarshalTo(o)
		switch v := o.GetClusterDiscoveryType().(type) {
		case *cluster.Cluster_Type:
			if v.Type != cluster.Cluster_EDS {
				return res
			}
		}
		key := ResourceKey{
			Name:    o.Name,
			TypeUrl: v3.EndpointType,
		}
		res = append(res, key)
	case v3.ListenerType:
		o := &listener.Listener{}
		_ = resp.Resource.UnmarshalTo(o)
		for _, fc := range getFilterChains(o) {
			for _, f := range fc.GetFilters() {
				if f.GetTypedConfig().GetTypeUrl() == "type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager" {
					hcm := &hcm.HttpConnectionManager{}
					_ = f.GetTypedConfig().UnmarshalTo(hcm)
					if r := hcm.GetRds().GetRouteConfigName(); r != "" {
						key := ResourceKey{
							Name:    r,
							TypeUrl: v3.RouteType,
						}
						res = append(res, key)
					}
				}
			}
		}
	}
	return res
}

func relate(parent, child *ResourceNode) {
	parent.Children[child] = struct{}{}
	child.Parents[parent] = struct{}{}
}

func (d *ADSC) DWatch() {
	log.Printf("sending initial watches")
	first := true
	for _, res := range d.initialWatches {
		req := &discovery.DeltaDiscoveryRequest{
			TypeUrl: res,
		}
		if first {
			req.Node = d.node
			first = false
		}
		err := d.dsend(req, ReasonInit)
		if err != nil {
			log.Printf("Error sending request: %v", err)
		}
	}
}

func (d *ADSC) dsend(dr *discovery.DeltaDiscoveryRequest, reason string) error {
	log.Printf("send message for type %v (%v) for +%v -%v", dr.TypeUrl, reason, dr.ResourceNamesSubscribe, dr.ResourceNamesUnsubscribe)
	return d.client.Send(dr)
}

func (d *ADSC) getNode(key ResourceKey) (*ResourceNode, bool) {
	found := true
	if d.tree[key] == nil {
		d.tree[key] = &ResourceNode{
			Key:      key,
			Parents:  map[*ResourceNode]struct{}{},
			Children: map[*ResourceNode]struct{}{},
		}
		found = false
	}
	return d.tree[key], found
}

func (d *ADSC) dumpTree() string {
	sb := strings.Builder{}
	roots := []*ResourceNode{}
	for _, n := range d.tree {
		if len(n.Parents) == 0 {
			roots = append(roots, n)
		}
	}
	for _, r := range roots {
		dumpNode(&sb, r, "")
	}
	return sb.String()
}

func (d *ADSC) deleteNode(node *ResourceNode, removals map[string][]string) {
	delete(d.tree, node.Key)
	for p := range node.Parents {
		delete(p.Children, node)
	}
	for c := range node.Children {
		delete(c.Parents, node)
		removals[c.Key.TypeUrl] = append(removals[c.Key.TypeUrl], c.Key.Name)
		if len(c.Parents) == 0 {
			d.deleteNode(c, removals)
		}
	}
}

func dumpNode(sb *strings.Builder, node *ResourceNode, indent string) {
	sb.WriteString(indent + node.Key.String() + ":\n")
	if len(indent) > 10 {
		return
	}
	for c := range node.Children {
		id := indent + "  "
		if _, f := c.Parents[node]; !f {
			id = indent + "**"
		}
		dumpNode(sb, c, id)
	}
}

func isWildcardTypeURL(typeURL string) bool {
	switch typeURL {
	case v3.SecretType, v3.EndpointType, v3.RouteType: // , v3.ExtensionConfigurationType:
		// By XDS spec, these are not wildcard
		return false
	case v3.ClusterType, v3.ListenerType:
		// By XDS spec, these are wildcard
		return true
	default:
		// All of our internal types use wildcard semantics
		return true
	}
}
