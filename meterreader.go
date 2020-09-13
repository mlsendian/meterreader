// meterreader, a cheap&nasty way to monitor a prepaid electrical meter
// Based on https://github.com/warthog618/gpio/blob/master/example/watcher/watcher.go

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	writeapi "github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/warthog618/gpio"
)

type Config struct {
	URL           string
	Key           string
	Organization  string
	Bucket        string
	Batchsize     uint
	FlushInterval uint
	Pin           int
}

// Grabbed from https://github.com/warthog618/gpio/blob/master/cmd/gppiio/gppiio.go
var pinNames = map[string]int{
	"J8P3":  gpio.J8p3,
	"J8P03": gpio.J8p3,
	"J8P5":  gpio.J8p5,
	"J8P05": gpio.J8p5,
	"J8P7":  gpio.J8p7,
	"J8P07": gpio.J8p7,
	"J8P8":  gpio.J8p8,
	"J8P08": gpio.J8p8,
	"J8P10": gpio.J8p10,
	"J8P11": gpio.J8p11,
	"J8P12": gpio.J8p12,
	"J8P13": gpio.J8p12,
	"J8P15": gpio.J8p15,
	"J8P16": gpio.J8p16,
	"J8P18": gpio.J8p18,
	"J8P19": gpio.J8p19,
	"J8P21": gpio.J8p21,
	"J8P22": gpio.J8p22,
	"J8P23": gpio.J8p23,
	"J8P24": gpio.J8p24,
	"J8P26": gpio.J8p26,
	"J8P27": gpio.J8p27,
	"J8P28": gpio.J8p28,
	"J8P29": gpio.J8p29,
	"J8P31": gpio.J8p31,
	"J8P32": gpio.J8p32,
	"J8P33": gpio.J8p33,
	"J8P35": gpio.J8p35,
	"J8P36": gpio.J8p36,
	"J8P37": gpio.J8p37,
	"J8P38": gpio.J8p38,
	"J8P40": gpio.J8p40,
}

var config Config

func init() {
	fmt.Println("meterreader v0.1.0, mls")

	viper.AutomaticEnv()
	viper.SetEnvPrefix("meterreader")

	pflag.String("influx_url", "", "Influx URL")
	pflag.String("influx_key", "", "Influx Key")
	pflag.String("influx_org", "", "Influx Org")
	pflag.String("influx_bucket", "electricity", "Influx Bucket")
	pflag.Uint("influx_batchsize", 20, "Influx batchsize")
	pflag.Uint("influx_flushinterval", 10000, "Influx flush interval in msecs")
	pflag.String("pinname", "J8P11", "Name of GPIO pin to use")

	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	config.URL = viper.GetString("influx_url")
	config.Key = viper.GetString("influx_key")
	config.Organization = viper.GetString("influx_org")
	config.Bucket = viper.GetString("influx_bucket")
	config.Batchsize = viper.GetUint("influx_batchsize")
	config.FlushInterval = viper.GetUint("influx_flushinterval")
	if config.URL == "" {
		fmt.Println("Missing the InfluxDB URL, set in the environment or use the flag")
		os.Exit(1)
	}
	if config.Key == "" {
		fmt.Println("Missing the InfluxDB key, set in the environment or use the flag")
		os.Exit(1)
	}
	fmt.Printf("Sending data to %s/%s/%s\n", config.URL, config.Organization, config.Bucket)

	var ok bool
	config.Pin, ok = pinNames[viper.GetString("pinname")]
	if !ok {
		fmt.Printf("Invalid pin name: %s\n", viper.GetString("pinname"))
		os.Exit(1)
	}
	fmt.Printf("Reading pin %d\n", config.Pin)
}

func setup_influx() writeapi.WriteAPI {
	options := influxdb2.DefaultOptions().SetBatchSize(config.Batchsize).SetFlushInterval(config.FlushInterval)
	client := influxdb2.NewClientWithOptions(config.URL, config.Key, options)
	writeAPI := client.WriteAPI(config.Organization, config.Bucket)
	errorsCh := writeAPI.Errors()
	go func() {
		for err := range errorsCh {
			fmt.Printf("write error: %s\n", err.Error())
		}
	}()

	return writeAPI
}

func setup_gpio_pin(writeAPI writeapi.WriteAPI) *gpio.Pin {
	// Setup GPIO
	err := gpio.Open()
	if err != nil {
		panic(err)
	}

	// Configure pin
	pin := gpio.NewPin(config.Pin)
	if err != nil {
		fmt.Printf("Unknown pin name %s", viper.GetString("pinname"))
	}
	pin.Input()
	pin.PullUp()

	// Setup interrupt handler for writing the LED pulse to Influx
	err = pin.Watch(gpio.EdgeFalling, func(pin *gpio.Pin) {
		p := influxdb2.NewPoint("power",
			map[string]string{"unit": "milliwatthour"},
			map[string]interface{}{"power": 1},
			time.Now())

		writeAPI.WritePoint(p)
	})
	if err != nil {
		panic(err)
	}
	return pin
}

func main() {
	// InfluxDB connection & error handling setup
	writeAPI := setup_influx()

	// Setup the GPIO interface, and close when we're done
	pin := setup_gpio_pin(writeAPI)
	defer gpio.Close()
	defer pin.Unwatch()

	// capture exit signals to ensure resources are released on exit.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	// wait for exit signal
	select {
	case <-quit:
	}
}
