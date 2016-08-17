package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"net"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/errors"
	"github.com/cerana/cerana/pkg/logrusx"
	"github.com/cerana/cerana/provider"
	"github.com/cerana/cerana/providers/dhcp"
	"github.com/krolaw/dhcp4"
	"github.com/pin/tftp"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type addrser interface {
	Addrs() ([]net.Addr, error)
}

type dhcpHandler struct {
	iface       addrser
	tracker     *acomm.Tracker
	coordinator *url.URL
}

var args = []string{
	"console=ttyS0",
	"cerana.boot_server=http://{{.IP}}/",
	"cerana.cluster_ips={{.IP}}",
	"cerana.mgmt_mac=${net0/mac}",
	"cerana.zfs_config=auto",
	"cerana.initrd_hash=sha256-{{.Hash}}",
}

var ipxe = template.Must(template.New("ipxe").Parse(`#!ipxe
kernel http://{{.IP}}/kernel ` + strings.Join(args, " ") + `
initrd http://{{.IP}}/initrd
boot
`))

func comm(tracker *acomm.Tracker, coordinator *url.URL, task string, args interface{}, resp interface{}) error {
	ch := make(chan *acomm.Response, 1)
	handler := func(_ *acomm.Request, resp *acomm.Response) {
		ch <- resp
	}

	req, err := acomm.NewRequest(acomm.RequestOptions{
		Task:           task,
		Args:           args,
		ResponseHook:   tracker.URL(),
		SuccessHandler: handler,
		ErrorHandler:   handler,
	})
	logrusx.DieOnError(err, "create request object")
	logrusx.DieOnError(tracker.TrackRequest(req, 0), "track request")
	logrusx.DieOnError(acomm.Send(coordinator, req), "send request")

	aResp := <-ch
	if aResp.Error != nil {
		return aResp.Error
	}

	logrusx.DieOnError(aResp.UnmarshalResult(resp), "deserialize result")
	return nil
}

func tftpReadHandler(undi []byte) func(string, io.ReaderFrom) error {
	fn := func(name string, rf io.ReaderFrom) error {
		if name != "undionly.kpxe" {
			return errors.Newv("unknown file", map[string]interface{}{"file": name})
		}

		_, err := rf.ReadFrom(bufio.NewReader(bytes.NewBuffer(undi)))
		return errors.Wrap(err)
	}

	return fn
}

func getFile(name string) ([]byte, string, time.Time) {
	errData := map[string]interface{}{"path": name}

	f, err := os.Open(name)
	logrusx.DieOnError(errors.Wrapv(err, errData), "open file")

	h := sha256.New()
	r := io.TeeReader(bufio.NewReader(f), h)
	buf, err := ioutil.ReadAll(r)
	logrusx.DieOnError(errors.Wrapv(err, errData), "read file")

	stat, err := f.Stat()
	logrusx.DieOnError(errors.Wrapv(err, errData), "stat file")
	_ = f.Close()

	return buf, fmt.Sprintf("%x", h.Sum(nil)), stat.ModTime()
}

func stringIP(ip net.IP) string {
	if ip == nil || ip.Equal(net.IPv4zero) {
		return ""
	}
	return ip.String()
}

func getIfaceIP(addrser addrser) net.IP {
	ips, err := addrser.Addrs()
	logrusx.DieOnError(errors.Wrap(err), "get interface addresses")

	if len(ips) < 1 {
		logrusx.DieOnError(errors.New("interface has no ip addresses configured"), "get interface addresses")
	}

	ip, _, err := net.ParseCIDR(ips[0].String())
	logrusx.DieOnError(errors.Wrapv(err, map[string]interface{}{"ip": ips[0].String()}), "parse ip address")

	return ip
}

func fillOptionsOffered(options dhcp4.Options, ip net.IP, offer dhcp.Lease) {
	dnses := make([]byte, len(offer.DNS)*4)
	for i, dns := range offer.DNS {
		copy(dnses[i*4:], []byte(dns.To4()))
	}
	options[dhcp4.OptionRouter] = []byte(offer.Gateway)
	options[dhcp4.OptionSubnetMask] = []byte(offer.Net.Mask)
	options[dhcp4.OptionDomainNameServer] = dnses

}

func nack(p dhcp4.Packet, ip net.IP) dhcp4.Packet {
	return dhcp4.ReplyPacket(p, dhcp4.NAK, ip, nil, 0, nil)
}

func (h *dhcpHandler) ServeDHCP(p dhcp4.Packet, msgType dhcp4.MessageType, options dhcp4.Options) (d dhcp4.Packet) {
	// TODO only handle known ciaddrs or known ciaddrs with us set as siaddr
	// aka w/e dnsmasq does

	reqIP := net.IP(options[dhcp4.OptionRequestedIPAddress])
	if reqIP == nil || reqIP.Equal(net.IPv4zero) {
		reqIP = p.CIAddr()
	}

	ip := getIfaceIP(h.iface).To4()

	sendOptions := dhcp4.Options{}
	userB, bootp := options[dhcp4.OptionUserClass]
	if bootp {
		switch string(userB) {
		case "iPXE":
			sendOptions[dhcp4.OptionBootFileName] = []byte(fmt.Sprintf("http://%s/boot.ipxe", ip))
			bootp = false
		default:
			sendOptions[dhcp4.OptionBootFileName] = []byte("undionly.kpxe")
		}
	}

	args := dhcp.Addresses{
		MAC: p.CHAddr().String(),
		IP:  stringIP(reqIP),
	}
	var offer dhcp.Lease
	var responseType dhcp4.MessageType
	switch msgType {
	case dhcp4.Discover:
		logrusx.DieOnError(comm(h.tracker, h.coordinator, "dhcp-offer-lease", args, &offer), "get lease offering")

		responseType = dhcp4.Offer
	case dhcp4.Request:
		server, ok := options[dhcp4.OptionServerIdentifier]
		if ok && !net.IP(server).Equal(ip) {
			// TODO only if CIADDR is unknown
			// Message not for this dhcp server
			return nil
		}

		if len(reqIP) != 4 {
			logrus.Error("requested ip is not IPv4")
			return nack(p, ip)
		} else if reqIP.Equal(net.IPv4zero) {
			logrus.Error("requested ip is 0.0.0.0")
			return nack(p, ip)
		} else if err := comm(h.tracker, h.coordinator, "dhcp-ack-lease", args, &offer); err != nil {
			logrus.WithField("error", err).Error("")
			return nack(p, ip)
		}

		responseType = dhcp4.ACK
	case dhcp4.Release, dhcp4.Decline:
		// just let it TTL timeout
		fallthrough
	default:
		return nil
	}

	fillOptionsOffered(sendOptions, ip, offer)
	responseOptions := sendOptions.SelectOrderOrAll(options[dhcp4.OptionParameterRequestList])
	reply := dhcp4.ReplyPacket(p, responseType, ip, offer.Net.IP, offer.Duration, responseOptions)
	if bootp {
		reply.SetSIAddr(ip)
		reply.PadToMinSize()
	}

	return reply
}

func main() {
	logrus.SetLevel(logrus.FatalLevel)

	v := viper.New()
	f := pflag.NewFlagSet("bootserver", pflag.ExitOnError)

	conf := newConfig(f, v)
	logrusx.DieOnError(f.Parse(os.Args), "parse argument")
	logrusx.DieOnError(conf.LoadConfig(), "load config")
	logrusx.DieOnError(conf.SetupLogging(), "setup logging")

	server, err := provider.NewServer(conf.Config)
	logrusx.DieOnError(err, "create server")

	tracker := server.Tracker()
	logrusx.DieOnError(tracker.Start(), "start tracker")

	iface, err := net.InterfaceByName(conf.iface())
	logrusx.DieOnError(errors.Wrapv(err, map[string]interface{}{"iface": conf.iface()}), "get interface")

	handler := &dhcpHandler{
		iface:       iface,
		coordinator: conf.CoordinatorURL(),
		tracker:     tracker,
	}

	dConn, err := net.ListenPacket("udp4", ":67")
	logrusx.DieOnError(errors.Wrapv(err, map[string]interface{}{"type": "udp4", "laddr": ":67"}), "bind dhcp port")

	undi, _, _ := getFile(conf.iPXE())
	tftpServer := tftp.NewServer(tftpReadHandler(undi), nil)
	tConn, err := net.ListenPacket("udp", ":69")
	logrusx.DieOnError(errors.Wrapv(err, map[string]interface{}{"type": "udp", "laddr": ":69"}), "bind tfp port")

	initrd, initrdHash, initrdMod := getFile(conf.initrd())

	buffer := &bytes.Buffer{}
	ipxeValues := map[string]string{
		"IP":   getIfaceIP(iface).String(),
		"Hash": initrdHash,
	}
	err = ipxe.Execute(buffer, ipxeValues)
	logrusx.DieOnError(errors.Wrapv(err, map[string]interface{}{"ip": ipxeValues["IP"], "hash": ipxeValues["Hash"]}), "generate ipxe boot script")
	bootScript := bytes.NewReader(buffer.Bytes())
	http.HandleFunc("/boot.ipxe", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		http.ServeContent(w, r, "", time.Time{}, bootScript)
	})

	kernel, _, kernelMod := getFile(conf.kernel())
	http.HandleFunc("/kernel", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		http.ServeContent(w, r, "", kernelMod, bytes.NewReader(kernel))
	})

	http.HandleFunc("/initrd", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		http.ServeContent(w, r, "", initrdMod, bytes.NewReader(initrd))
	})

	hConn, err := net.Listen("tcp", ":80")
	logrusx.DieOnError(errors.Wrapv(err, map[string]interface{}{"type": "tcp", "laddr": ":80"}), "bind http port")

	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		logrus.Info("serving tftp")
		tftpServer.Serve(tConn.(*net.UDPConn))
		wg.Done()
	}()
	go func() {
		logrus.Info("serving http")
		if err := http.Serve(hConn, nil); err != nil {
			logrus.WithField("error", errors.Wrap(err)).Error("http error")
		}
		wg.Done()
	}()
	go func() {
		logrus.Info("serving dhcp")
		if err := dhcp4.ServeIf(iface.Index, dConn, handler); err != nil {
			logrus.WithField("error", errors.Wrap(err)).Error("dhcp error")
		}
		wg.Done()
	}()
	wg.Wait()
}
