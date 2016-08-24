package main

import (
	"strconv"

	"github.com/cerana/cerana/pkg/errors"
	"github.com/cerana/cerana/provider"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type config struct {
	*provider.Config
	viper *viper.Viper
}

func newConfig(f *pflag.FlagSet, v *viper.Viper) config {
	f.String("dhcp-interface", "", "interface to bind dhcp server to")
	f.String("ipxe", "undionly.pxe", "iPXE file to serve to non-iPXE clients")
	f.String("kernel", "kernel", "kernel file")
	f.String("initrd", "initrd", "initrd file")
	f.Uint16("dhcp-port", 67, "port to listen for DHCP connections")
	f.Uint16("http-port", 80, "port to listen for HTTP connections")
	f.Uint16("tftp-port", 69, "port to listen for TFTP connections")
	return config{
		Config: provider.NewConfig(f, v),
		viper:  v,
	}
}

func (c config) LoadConfig() error {
	if err := c.Config.LoadConfig(); err != nil {
		return err
	}

	return c.validate()
}

func (c config) validate() error {
	if c.viper.GetString("dhcp-interface") == "" {
		return errors.New("interface is not defined")
	}
	return c.Config.Validate()
}

func (c config) iface() string {
	return c.viper.GetString("dhcp-interface")
}

func (c config) iPXE() string {
	return c.viper.GetString("ipxe")
}

func (c config) kernel() string {
	return c.viper.GetString("kernel")
}

func (c config) initrd() string {
	return c.viper.GetString("initrd")
}

func (c config) dhcp() string {
	return strconv.Itoa(c.viper.GetInt("dhcp-port"))
}

func (c config) http() string {
	return strconv.Itoa(c.viper.GetInt("http-port"))
}

func (c config) tftp() string {
	return strconv.Itoa(c.viper.GetInt("tftp-port"))
}
