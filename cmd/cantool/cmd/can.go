package cmd

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/roffe/gocan"
	"github.com/roffe/gocan/adapter/lawicel"
	"github.com/roffe/gocan/adapter/obdlink"
	"github.com/roffe/gocan/pkg/ecu"
	"github.com/spf13/cobra"
)

var canCMD = &cobra.Command{
	Use:   "can",
	Short: "CAN related commands",
	Args:  cobra.ExactArgs(1),
}

func init() {
	rootCmd.AddCommand(canCMD)
}
func initCAN(ctx context.Context) (*gocan.Client, error) {
	adapter, port, baudrate, canrate, err := getAdapterOpts()
	if err != nil {
		return nil, err
	}

	var dev gocan.Adapter
	switch strings.ToLower(adapter) {
	case "canusb":
		dev = lawicel.NewCanusb()
	case "sx", "obdlinksx":
		dev = obdlink.NewSX()
	default:
		return nil, fmt.Errorf("unknown adapter %q", adapter)
	}

	if err := dev.SetPort(port); err != nil {
		return nil, err
	}
	if err := dev.SetPortRate(baudrate); err != nil {
		return nil, err
	}
	if err := dev.SetCANrate(canrate); err != nil {
		return nil, err
	}

	switch getECUType() {
	case ecu.Trionic5:
		dev.SetCANfilter(0x000, 0x005, 0x006, 0x00C)
	case ecu.Trionic7:
		dev.SetCANfilter(0x220, 0x238, 0x240, 0x258, 0x266)
	case ecu.Trionic8:
		dev.SetCANfilter(0x011, 0x311, 0x7E0, 0x7E8, 0x5E8)
	}

	if err := dev.Init(ctx); err != nil {
		return nil, err
	}

	return gocan.New(ctx, dev, filters)
}

func getECUType() ecu.Type {
	ecutemp, err := rootCmd.PersistentFlags().GetString(flagECUType)
	if err != nil {
		log.Fatal(err)
	}
	return ecu.TypeFromString(ecutemp)
}

func getAdapterOpts() (adapter string, port string, baudrate int, canrate float64, err error) {
	pf := rootCmd.PersistentFlags()

	port, err = pf.GetString(flagPort)
	if err != nil {
		return
	}
	baudrate, err = pf.GetInt(flagBaudrate)
	if err != nil {
		return
	}
	adapter, err = pf.GetString(flagAdapter)
	if err != nil {
		return
	}
	canRate, err := pf.GetString(flagCANRate)
	if err != nil {
		return
	}

	switch strings.ToLower(canRate) {
	case "pbus":
		canrate = 500
	case "ibus":
		canrate = 47.619
	case "t5":
		canrate = 615.384
	default:
		f, errs := strconv.ParseFloat(canRate, 64)
		if errs != nil {
			err = fmt.Errorf("invalid CAN rate: %q: %v", canRate, err)
			return
		}
		canrate = f
	}
	return
}
