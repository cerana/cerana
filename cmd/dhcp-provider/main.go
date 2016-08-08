package main

import (
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/logrusx"
	"github.com/cerana/cerana/provider"
	"github.com/cerana/cerana/providers/clusterconf"
	"github.com/cerana/cerana/providers/dhcp"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

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
	logrusx.DieOnError(err, "create request object")
	logrusx.DieOnError(tracker.TrackRequest(req, 0), "track request")

	if err = acomm.Send(coordinator, req); err != nil {
		return err
	}

	aResp := <-ch
	if aResp.Error != nil {
		return aResp.Error
	}

	logrusx.DieOnError(aResp.UnmarshalResult(resp), "deserialize result")
	return nil
}

func getDHCPConfig(tracker *acomm.Tracker, coordinator *url.URL) (*clusterconf.DHCPConfig, error) {
	dconf := &clusterconf.DHCPConfig{}
	err := comm(tracker, coordinator, "get-dhcp-config", nil, dconf)
	logrusx.DieOnError(err, "get dhcp configuration")
	return dconf
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

	config := dhcp.NewConfig(f, v)
	logrusx.DieOnError(f.Parse(os.Args), "parse arguments")
	logrusx.DieOnError(config.LoadConfig(), "load configuration")
	logrusx.DieOnError(config.SetupLogging(), "setup logging")

	server, err := provider.NewServer(config.Config)
	logrusx.DieOnError(err, "create provider")
	logrusx.DieOnError(server.Tracker().Start(), "start tracker")

	storedConfig := getDHCPConfig(server.Tracker(), config.CoordinatorURL())

	v.Set("dns-servers", storedConfig.DNS)
	v.Set("gateway", storedConfig.Gateway)
	v.Set("lease-duration", storedConfig.Duration)
	v.Set("network", storedConfig.Net)

	d, err := dhcp.New(config, server.Tracker())
	logrusx.DieOnError(err, "create dhcp server")

	logrus.WithField("config", storedConfig).Info("starting provider")

	d.RegisterTasks(server)
	if len(server.RegisteredTasks()) == 0 {
		logrus.Warn("no registered tasks, exiting")
		os.Exit(1)
	}
	logrusx.DieOnError(server.Start(), "successfully run dhcp provider")
	server.StopOnSignal()
}
