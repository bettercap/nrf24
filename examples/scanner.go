package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/bettercap/nrf24"
)

var (
	hopPeriod = 100 * time.Millisecond
)

func address(raw []byte) string {
	parts := []string{}
	for _, b := range raw {
		parts = append(parts, fmt.Sprintf("%X", b))
	}
	return strings.Join(parts, ":")
}

func main() {
	fmt.Printf("nRF24LU1+ - RFStorm Scanner\n\n")

	dongle, err := nrf24.Open()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	defer dongle.Close()

	fmt.Printf("device open: %s\n", dongle.String())

	if err = dongle.EnterPromiscMode(); err != nil {
		fmt.Printf("error: %v\n", err)
		return
	} else {
		fmt.Printf("device is in promisc mode\n\n")
	}

	ch := 1
	chIndex := 0
	lastHop := time.Time{}

	for {
		if time.Since(lastHop) >= hopPeriod {
			chIndex++
			ch = chIndex % nrf24.TopChannel
			if err := dongle.SetChannel(ch); err != nil {
				fmt.Printf("error setting channel %d: %v\n", ch, err)
			} else {
				lastHop = time.Now()
			}
		}

		if buf, err := dongle.ReceivePayload(); err != nil {
			fmt.Printf("error receiving payload on channel %d: %v\n", ch, err)
		} else if len(buf) >= 5 {
			addr, payload := buf[0:5], buf[5:]
			fmt.Printf("found device %s on channel %d (payload:%x)\n", address(addr), ch, payload)
		}
	}
}
