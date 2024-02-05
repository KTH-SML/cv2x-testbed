package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/beevik/ntp"
)

type NTPClient struct {
	Url  string
	Resp ntp.Response
}

func ConnectNTP(url string) (*NTPClient, error) {
	if len(url) == 0 {
		return nil, fmt.Errorf("NTP URL is empty")
	}
	return &NTPClient{Url: url}, nil
}

func (c *NTPClient) GetOffset() int64 {
	return c.Resp.ClockOffset.Nanoseconds()
}

func (c *NTPClient) SingleQuery() {
	resp, err := ntp.Query(c.Url)
	if err != nil {
		fmt.Printf("NTP error: %v\n", err)
	} else {
		c.Resp = *resp
	}
}

var ntpMaxPoll = flag.Duration("ntpMaxPoll", 250*time.Millisecond, "NTP Max Poll Interval")

func (c *NTPClient) QueryLoop(pred func(ntp.Response) bool) {
	for pred(c.Resp) {
		c.SingleQuery()
		sleepDur := *ntpMaxPoll
		// Poll is the maximum interval between successive NTP polling messages. It is not relevant for simple NTP clients like this one.
		if c.Resp.Poll < sleepDur {
			// sleepDur = c.Resp.Poll
		}
		time.Sleep(sleepDur)
	}
}
