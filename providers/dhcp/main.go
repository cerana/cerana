package dhcp

import (
	"errors"
	"math/rand"
	"net"
	"net/url"
	"sort"
	"time"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/provider"
	"github.com/cerana/cerana/providers/kv"
	"github.com/krolaw/dhcp4"
)

const prefix string = "dhcp-leases/"

var ttlOffer = 1 * time.Minute

// DHCP is a provider of dhcp functionality.
type DHCP struct {
	coordinator *url.URL
	tracker     *acomm.Tracker
	config      *Config
	maxIP       net.IP
	randIP      func() net.IP
}

// Addresses specifies the argument to all endpoints
type Addresses struct {
	MAC string `json:"mac"`
	IP  string `json:"ip"`
}

// Lease specifies the dhcp lease returned from the "dhcp-offer-lease" endpoint.
type Lease struct {
	DNS      []net.IP      `json:"dns"`
	Duration time.Duration `json:"duration"`
	Gateway  net.IP        `json:"gateway"`
	Net      net.IPNet     `json:"net"`
}

// New creates a new instance of DHCP.
func New(config *Config, tracker *acomm.Tracker) (*DHCP, error) {
	network := config.Network()
	ones, bits := network.Mask.Size()
	size := 1<<uint(bits-ones) - 2
	maxIP := dhcp4.IPAdd(network.IP, size+2)

	return &DHCP{
		coordinator: config.CoordinatorURL(),
		tracker:     tracker,
		config:      config,
		maxIP:       maxIP,
		randIP: func() net.IP {
			num := rand.Intn(size) + 1
			return dhcp4.IPAdd(network.IP, num)
		},
	}, nil
}

// RegisterTasks registers all of DHCP's task handlers with the server.
func (d *DHCP) RegisterTasks(server *provider.Server) {
	server.RegisterTask("dhcp-offer-lease", d.get)
	server.RegisterTask("dhcp-ack-lease", d.ack)
}

func lookupMAC(tracker *acomm.Tracker, coord *url.URL, ip string) (string, error) {
	ch := make(chan *acomm.Response, 1)
	handler := func(_ *acomm.Request, resp *acomm.Response) {
		ch <- resp
	}
	req, err := acomm.NewRequest(acomm.RequestOptions{
		Task:           "kv-get",
		ResponseHook:   tracker.URL(),
		SuccessHandler: handler,
		ErrorHandler:   handler,
		Args:           kv.GetArgs{Key: prefix + ip},
	})
	if err != nil {
		return "", err
	}

	if err = tracker.TrackRequest(req, 0); err != nil {
		return "", err
	}

	if err = acomm.Send(coord, req); err != nil {
		return "", err
	}

	resp := <-ch
	if resp.Error != nil {
		return "", nil
	}

	kvp := kv.Value{}
	if err = resp.UnmarshalResult(&kvp); err != nil {
		return "", err
	}

	return string(kvp.Data), nil
}

func doESet(tracker *acomm.Tracker, coord *url.URL, mac, ip string, ttl time.Duration) (chan *acomm.Response, error) {
	ch := make(chan *acomm.Response, 1)
	handler := func(_ *acomm.Request, resp *acomm.Response) {
		ch <- resp
	}
	req, err := acomm.NewRequest(acomm.RequestOptions{
		Task:           "kv-ephemeral-set",
		ResponseHook:   tracker.URL(),
		SuccessHandler: handler,
		ErrorHandler:   handler,
		Args: kv.EphemeralSetArgs{
			Key:   prefix + ip,
			Value: mac,
			TTL:   ttl,
		},
	})
	if err != nil {
		return nil, err
	}

	if err = tracker.TrackRequest(req, 0); err != nil {
		return nil, err
	}

	if err = acomm.Send(coord, req); err != nil {
		return nil, err
	}

	return ch, nil
}

func refreshLeasePending(tracker *acomm.Tracker, coord *url.URL, mac, ip string) (bool, error) {
	ch, err := doESet(tracker, coord, mac, ip, ttlOffer)
	if err != nil {
		return false, err
	}

	// we don't care about kv returning an error per se, this just means that caller will try a different ip address
	if resp := <-ch; resp.Error != nil {
		return false, nil
	}

	return true, nil
}

func refreshLeaseAck(tracker *acomm.Tracker, coord *url.URL, mac, ip string, ttl time.Duration) error {
	ch, err := doESet(tracker, coord, mac, ip, ttl)
	if err != nil {
		return err
	}

	resp := <-ch
	return resp.Error
}

func nextGetter(closer <-chan struct{}, taken []uint32, min, max uint32) <-chan uint32 {
	ch := make(chan uint32)
	next := min
	go func() {
		for {
			next++
			// make sure we stop before max
			if next == max {
				close(ch)
				return
			}

			// check if ip in in the taken list, if so pop it from the list
			if len(taken) > 0 {
				if next == taken[0] {
					taken = taken[1:]
					continue
				}
			}

			select {
			case ch <- next:
			case <-closer:
				return
			}
		}
	}()

	return ch
}

// findHole will find a skipped number or last+1
// slice must already be sorted
func findHole(slice []uint32) uint32 {
	if len(slice) == 0 {
		return 0
	}
	hole := slice[0] + 1
	for i := range slice {
		u := slice[i]
		if u > hole {
			break
		}
		hole = u + 1
	}
	return hole
}

func getAllAllocations(tracker *acomm.Tracker, coord *url.URL) (map[string]string, error) {
	ch := make(chan *acomm.Response, 1)
	handler := func(_ *acomm.Request, resp *acomm.Response) {
		ch <- resp
	}
	req, err := acomm.NewRequest(acomm.RequestOptions{
		Task:           "kv-getAll",
		ResponseHook:   tracker.URL(),
		SuccessHandler: handler,
		ErrorHandler:   handler,
		Args:           kv.GetArgs{Key: prefix},
	})
	if err != nil {
		return nil, err
	}

	if err = tracker.TrackRequest(req, 0); err != nil {
		return nil, err
	}

	if err = acomm.Send(coord, req); err != nil {
		return nil, err
	}

	resp := <-ch
	if resp.Error != nil {
		return nil, err
	}

	kvs := map[string]kv.Value{}
	if err = resp.UnmarshalResult(&kvs); err != nil {
		return nil, err
	}

	allocations := make(map[string]string, len(kvs))
	for k, v := range kvs {
		allocations[k[len(prefix):]] = string(v.Data)
	}
	return allocations, nil
}

// get reserves an ip address upon client dhcp request
func (d *DHCP) get(req *acomm.Request) (interface{}, *url.URL, error) {
	addrs := Addresses{}
	if err := req.UnmarshalArgs(&addrs); err != nil {
		return nil, nil, err
	}
	if addrs.MAC == "" {
		return nil, nil, errors.New("missing arg: mac")
	}

	lease := Lease{
		DNS:      d.config.DNSServers(),
		Duration: d.config.LeaseDuration(),
		Gateway:  d.config.Gateway(),
		Net: net.IPNet{
			Mask: d.config.Network().Mask,
		},
	}

	// shortcut client renewing ip
	if addrs.IP != "" {
		mac, err := lookupMAC(d.tracker, d.coordinator, addrs.IP)
		if err != nil {
			return nil, nil, err
		}
		if mac == addrs.MAC {
			ok, err := refreshLeasePending(d.tracker, d.coordinator, addrs.MAC, addrs.IP)
			if err != nil {
				return nil, nil, err
			}
			if ok {
				lease.Net.IP = net.ParseIP(addrs.IP)
				return lease, nil, nil
			}
		}
		addrs.IP = ""
	}

	// no ip or the requested ip has been recycled already
	// TODO(mmlb) be more clever
	// we could probably start building up a cache of mac:ip that we can verify in the kv
	// but if we really wanted to be clever/efficient then we would add transactions to pkg/kv and have /dhcp/ips and /dhcp/macs

	leases, err := getAllAllocations(d.tracker, d.coordinator)
	if err != nil {
		return nil, nil, err
	}

	ips := uIPs(make([]uint32, 0, len(leases)))
	for ip, mac := range leases {
		ips = append(ips, ipToU32(net.ParseIP(ip)))
		if mac == addrs.MAC {
			addrs.IP = ip
		}
	}

	if addrs.IP != "" {
		ok, err := refreshLeasePending(d.tracker, d.coordinator, addrs.MAC, addrs.IP)
		if err != nil {
			return nil, nil, err
		}
		if ok {
			lease.Net.IP = net.ParseIP(addrs.IP)
			return lease, nil, nil
		}
	}

	for range [5]struct{}{} {
		ip := d.randIP()
		if !d.config.ipInRange(ip) {
			continue
		}
		if _, ok := leases[ip.String()]; ok {
			continue
		}

		addrs.IP = ip.String()
		ok, err := refreshLeasePending(d.tracker, d.coordinator, addrs.MAC, addrs.IP)
		if err != nil {
			return nil, nil, err
		}
		if ok {
			lease.Net.IP = net.ParseIP(addrs.IP)
			return lease, nil, nil
		}
	}

	// so that didn't work, now we actually have to _find_ a lease

	// so go through the slice of allocations and try to reserve one of the unallocated ips
	// if we go through all of ips in the subnet then fail
	sort.Sort(ips)

	if len(ips) == 0 || ips[0] > ipToU32(d.config.Network().IP)+1 {
		ip := dhcp4.IPAdd(d.config.Network().IP, 1)
		_, err := refreshLeasePending(d.tracker, d.coordinator, addrs.MAC, ip.String())
		if err != nil {
			return nil, nil, err
		}
		lease.Net.IP = ip
		return lease, nil, nil
	}

	closer := make(chan struct{}, 1)
	defer close(closer)
	for uIP := range nextGetter(closer, ips, ipToU32(d.config.Network().IP), ipToU32(d.maxIP)) {
		ip := u32ToIP(uIP)

		addrs.IP = ip.String()
		ok, err := refreshLeasePending(d.tracker, d.coordinator, addrs.MAC, addrs.IP)
		if err != nil {
			return nil, nil, err
		}

		if ok {
			lease.Net.IP = ip
			return lease, nil, nil
		}
	}

	return nil, nil, errors.New("no IP address available")
}

func (d *DHCP) ack(req *acomm.Request) (interface{}, *url.URL, error) {
	addrs := Addresses{}
	if err := req.UnmarshalArgs(&addrs); err != nil {
		return nil, nil, err
	}
	if addrs.MAC == "" {
		return nil, nil, errors.New("missing arg: mac")
	}
	if addrs.IP == "" {
		return nil, nil, errors.New("missing arg: ip")
	}

	mac, err := lookupMAC(d.tracker, d.coordinator, addrs.IP)
	if err != nil {
		return nil, nil, err
	}
	if addrs.MAC != mac {
		return nil, nil, errors.New("requested ip not assigned to this mac")
	}

	return nil, nil, refreshLeaseAck(d.tracker, d.coordinator, addrs.MAC, addrs.IP, d.config.LeaseDuration())
}
