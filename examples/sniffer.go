package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/bettercap/nrf24"
)

var (
	pingPeriod  = 100 * time.Millisecond
	address     = ""
	rawAddress  = []byte(nil)
	pingPayload = []byte{0x0f, 0x0f, 0x0f, 0x0f}
	err         = error(nil)
)

func init() {
	flag.StringVar(&address, "address", "", "Address to sniff for.")
	flag.Parse()
}

func convertAddress() (error, []byte) {
	if address == "" {
		return fmt.Errorf("no --address specified"), nil
	}

	clean := strings.Replace(address, ":", "", -1)
	raw, err := hex.DecodeString(clean)
	if err != nil {
		return err, nil
	} else if len(raw) != 5 {
		return fmt.Errorf("address must be composed of 5 octets"), nil
	}
	// https://github.com/golang/go/wiki/SliceTricks
	for i := len(raw)/2 - 1; i >= 0; i-- {
		opp := len(raw) - 1 - i
		raw[i], raw[opp] = raw[opp], raw[i]
	}

	return nil, raw
}

func main() {
	fmt.Printf("nRF24LU1+ - RFStorm Sniffer\n\n")

	if err, rawAddress = convertAddress(); err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	dongle, err := nrf24.Open()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	defer dongle.Close()

	fmt.Printf("device open: %s\n", dongle.String())

	if err = dongle.EnterSnifferModeFor(rawAddress); err != nil {
		fmt.Printf("error: %v\n", err)
		return
	} else {
		fmt.Printf("device is in sniffer mode\n\n")
	}

	lastPing := time.Time{}
	ch, _ := dongle.GetChannel()

	for {
		if time.Since(lastPing) >= pingPeriod {
			pinged := false
			if err = dongle.TransmitPayload(pingPayload, 250, 1); err != nil {
				for ch = 1; ch <= nrf24.TopChannel && !pinged; ch++ {
					if err := dongle.SetChannel(ch); err != nil {
						fmt.Printf("error setting channel %d: %v\n", ch, err)
					} else if err = dongle.TransmitPayload(pingPayload, 250, 1); err == nil {
						pinged = true
						lastPing = time.Now()
						break
					}
				}
			}
		}

		if buf, err := dongle.ReceivePayload(); err != nil {
			fmt.Printf("error receiving payload: %v\n", err)
		} else if buf[0] == 0 {
			buf = buf[1:]
			fmt.Printf("[%s] (channel %02d) : %x\n", address, ch, buf)
		}
	}
}
