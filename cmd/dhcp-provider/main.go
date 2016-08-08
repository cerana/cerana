package main

import (
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/logrusx"
	"github.com/cerana/cerana/provider"
	"github.com/cerana/cerana/providers/clusterconf"
	"github.com/cerana/cerana/providers/dhcp"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func dieOnError(msg string, err error) {
	if err != nil {
		logrus.WithError(err).Fatal("failed to " + msg)
	}
}

func defaultNetwork() net.IPNet {
	_, net, _ := net.ParseCIDR("172.16.0.0/16")
	return *net
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
	if err = acomm.Send(coordinator, req); err != nil {
		return err
	}

	aResp := <-ch
	if aResp.Error != nil {
		return aResp.Error
	}

	dieOnError("deserialize result", aResp.UnmarshalResult(resp))
	return nil
}

func getDHCPConfig(tracker *acomm.Tracker, coordinator *url.URL) (*clusterconf.DHCPConfig, error) {
	dconf := &clusterconf.DHCPConfig{}
	return dconf, comm(tracker, coordinator, "get-dhcp", nil, dconf)
}

func setDHCPConfig(tracker *acomm.Tracker, coordinator *url.URL, config *dhcp.Config) {
	dconf := clusterconf.DHCPConfig{}
	err := comm(tracker, coordinator, "get-dhcp", nil, &dconf)
	if err == nil {
		logrus.Warn("dhcp configuration exists, not overriding it")
		return
	}

	dconf = clusterconf.DHCPConfig{
		DNS:      config.DNSServers(),
		Duration: config.LeaseDuration(),
		Gateway:  config.Gateway(),
		Net:      *config.Network(),
	}
	err = comm(tracker, coordinator, "set-dhcp", dconf, nil)
	dieOnError("set dhcp configuration", err)
}

func joinDNS(dns []net.IP) string {
	strs := make([]string, len(dns))
	for i := range dns {
		strs[i] = dns[i].String()
	}
	return strings.Join(strs, ",")
}

func main() {
	logrus.SetFormatter(&logrusx.JSONFormatter{})

	v := viper.New()
	f := pflag.NewFlagSet("dhcp-provider", pflag.ExitOnError)
	f.String("dns-servers", "", "[optional] comma separated list of dns servers ")
	f.IP("gateway", nil, "[optional] default gateway")
	f.Duration("lease-duration", 24*time.Hour, "default lease duration")
	f.IPNet("network", defaultNetwork(), "network to manage dhcp addresses on")

	config := dhcp.NewConfig(f, v)
	dieOnError("parse arguments", f.Parse(os.Args))
	dieOnError("load configuration", config.LoadConfig())
	dieOnError("setup logging", config.SetupLogging())

	set := v.IsSet("dns-servers") || v.IsSet("gateway") || v.IsSet("lease-duration") || v.IsSet("network")

	server, err := provider.NewServer(config.Config)
	dieOnError("create provider", err)
	dieOnError("start tracker", server.Tracker().Start())

	if set == true {
		setDHCPConfig(server.Tracker(), config.CoordinatorURL(), config)
	}
	var storedConfig *clusterconf.DHCPConfig
	for {
		storedConfig, err = getDHCPConfig(server.Tracker(), config.CoordinatorURL())
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	v.Set("dns-servers", joinDNS(storedConfig.DNS))
	v.Set("gateway", storedConfig.Gateway)
	v.Set("lease-duration", storedConfig.Duration)
	v.Set("network", storedConfig.Net.String())

	d, err := dhcp.New(config, server.Tracker())
	dieOnError("create dhcp server", err)

	logrus.WithField("config", storedConfig).Info("starting provider")

	d.RegisterTasks(server)
	if len(server.RegisteredTasks()) == 0 {
		logrus.Warn("no registered tasks, exiting")
		os.Exit(1)
	}
	dieOnError("successfully run dhcp provider", server.Start())
	server.StopOnSignal()
}
