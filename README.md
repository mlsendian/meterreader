# Electrical Meter Monitor

Stupidly simple project for tracking a prepaid electrical meter that has a flashing status LED.

The tiny custom board triggers a signal on a Raspberry PI Zero W's GPIO pin, which writes a point to an InfluxDB.

# Board

The stupidly simple circuit is provided in a Fritzing file. With an LDR, LM358 and a few other resistors, fashion a housing for the PDR and stick over the status LED.

# Built with

 - Go 1.15 on OSX

# Building & Installation

```
$ git clone https://github.com/mlsendian/meterreader
$ cd meterreader
$ make
# Edit Makefile and change RASPBERRYPI to the IP/hostname of your local Raspi
$ make ssh_install
```

# Configuration

After installation, SSH to the Raspi and edit `/etc/default/meterreader`. At the minimum you must set these values:
 - `METERREADER_INFLUX_KEY` - an InfluxDB API key
 - `METERREADER_INFLUX_URL` - an InfluxDB URL (e.g. http://192.168.1.1:9999)
 - `METERREADER_INFLUX_ORG` - InfluxDB organization to write to

Other possible config values are:
 - `METERREADER_INFLUX_BATCHSIZE` - Max number of events to batch together (default: 20)
 - `METERREADER_INFLUX_BUCKET` - InfluxDB bucket to write to (default: "electricity")
 - `METERREADER_INFLUX_FLUSHINTERVAL` - Max time period in msecs to rollup events in (default: 10000)
 - `METERREADER_PINNAME` - Name of Raspi pin to use, see `meterreader.go` for others (default: "J8P11")

All config values can also be specified on the command line, see `meterreader --help`.