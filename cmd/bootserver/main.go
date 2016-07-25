package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"errors"
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

func dieOnError(msg string, err error) {
	if err != nil {
		logrus.WithError(err).Fatal("failed to " + msg)
	}
}

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
	dieOnError("create request object", err)
	dieOnError("track request", tracker.TrackRequest(req, 0))
	dieOnError("send request", acomm.Send(coordinator, req))

	aResp := <-ch
	if aResp.Error != nil {
		return aResp.Error
	}

	dieOnError("deserialize result", aResp.UnmarshalResult(resp))
	return nil
}

func tftpReadHandler(undi []byte) func(string, io.ReaderFrom) error {
	fn := func(name string, rf io.ReaderFrom) error {
		if name != "undionly.kpxe" {
			logrus.WithError(errors.New("unknown file")).WithField("file", name).Error("")
		}

		_, err := rf.ReadFrom(bufio.NewReader(bytes.NewBuffer(undi)))
		if err != nil {
			logrus.WithError(err).Error("error sending iPXE image")
		}
		return err
	}

	return fn
}

func getFile(name string) ([]byte, string, time.Time) {
	f, err := os.Open(name)
	dieOnError("open file", err)

	h := sha256.New()
	r := io.TeeReader(bufio.NewReader(f), h)
	buf, err := ioutil.ReadAll(r)
	dieOnError("read file", err)

	stat, err := f.Stat()
	dieOnError("stat file", err)
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
	dieOnError("get interface addresses", err)

	if len(ips) < 1 {
		dieOnError("get interface addresses", errors.New("interface has no ip addresses configured"))
	}

	ip, _, err := net.ParseCIDR(ips[0].String())
	dieOnError("parse ip address", err)

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
		dieOnError("get lease offering", comm(h.tracker, h.coordinator, "dhcp-offer-lease", args, &offer))

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
			logrus.WithError(err).Error("")
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
	dieOnError("parse argument", f.Parse(os.Args))
	dieOnError("load config", conf.LoadConfig())
	dieOnError("setup logging", conf.SetupLogging())

	server, err := provider.NewServer(conf.Config)
	dieOnError("create server", err)

	tracker := server.Tracker()
	dieOnError("start tracker", tracker.Start())

	iface, err := net.InterfaceByName(conf.iface())
	dieOnError("get interface", err)

	handler := &dhcpHandler{
		iface:       iface,
		coordinator: conf.CoordinatorURL(),
		tracker:     tracker,
	}

	dConn, err := net.ListenPacket("udp4", ":67")
	dieOnError("bind dhcp port", err)

	undi, _, _ := getFile(conf.iPXE())
	tftpServer := tftp.NewServer(tftpReadHandler(undi), nil)
	tConn, err := net.ListenPacket("udp", ":69")
	dieOnError("bind tfp port", err)

	initrd, initrdHash, initrdMod := getFile(conf.initrd())

	buffer := &bytes.Buffer{}
	err = ipxe.Execute(buffer, map[string]string{
		"IP":   getIfaceIP(iface).String(),
		"Hash": initrdHash,
	})
	dieOnError("generate ipxe boot script", err)
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
	dieOnError("bind http port", err)

	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		logrus.Info("serving tftp")
		tftpServer.Serve(tConn.(*net.UDPConn))
		wg.Done()
	}()
	go func() {
		logrus.Info("serving http")
		err := http.Serve(hConn, nil)
		logrus.Info("http err:", err)
		wg.Done()
	}()
	go func() {
		logrus.Info("serving dhcp")
		err := dhcp4.ServeIf(iface.Index, dConn, handler)
		logrus.Info("dhcp err:", err)
		wg.Done()
	}()
	wg.Wait()
}
