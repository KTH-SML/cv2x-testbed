package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/beevik/ntp"
	"github.com/bluenviron/goroslib/v2"
	"github.com/bluenviron/goroslib/v2/pkg/msgs/sensor_msgs"
	"github.com/nats-io/nats.go"
)

func connect(host string) *nats.EncodedConn {

	nc, err := nats.Connect(host, nats.Timeout(1*time.Minute))
	if err != nil {
		panic(err)
	}
	c, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		panic(err)
	}
	return c
}

var nodeName = flag.String("name", "", "Name of node, defaults to same name as type.")
var nodeType = flag.String("type", "", "Type of node, e.g. server")
var natsAddr = flag.String("host", "10.20.33.130", "URL to NATS server host.")
var ntpAddr = flag.String("ntp", "10.47.6.47", "URL to NTP server.")
var enableROS = flag.Bool("ros", false, "Enable ROS.")

func main() {
	flag.Parse()

	if len(*nodeName) == 0 {
		*nodeName = *nodeType
	}

	fmt.Printf("Starting %s!\n", *nodeName)

	natsClient := connect(*natsAddr)

	ntpClient, err := ConnectNTP(*ntpAddr)
	if err != nil {
		log.Fatal(err)
	}

	var node *Node
	if *nodeType == "coordinator" {
		node = NewNode(*nodeName, natsClient, ntpClient, coordinator())
		node.attr["paused"] = 0
	} else if *nodeType == "sensor" {
		node = NewNode(*nodeName, natsClient, ntpClient, func(node *Node) {
			size, _ := node.attr["DATA_SIZE"]
			seq, _ := node.attr["DATA_SEQ"]
			data, err := RandomBytes(size)
			if err != nil {
				panic(err)
			}

			message := Packet{}
			message.Header.Stamp = time.Now().UnixNano()
			message.Header.Seq = int64(seq)
			message.Data = data
			message.Chk = Checksum(data, 0)
			message.T1 = time.Now().UnixNano()
			message.E1 = node.ntpClient.Resp.ClockOffset.Nanoseconds()
			node.nc.Publish(fmt.Sprintf("%s.data", node.name), &message)

			node.attr["DATA_SEQ"] += 1
		})
		node.attr["DATA_SIZE"] = 1000
		node.attr["DATA_SEQ"] = 0
	} else if *nodeType == "server" {
		node = NewNode(*nodeName, natsClient, ntpClient, func(node *Node) {})
		node.attr["COMPUTE_TIME"] = 0
		node.nc.Subscribe("sensor.data", func(p *Packet) {
			p.T2 = time.Now().UnixNano()
			p.E2 = node.ntpClient.GetOffset()
			if c_time, ok := node.attr["COMPUTE_TIME"]; ok {
				time.Sleep(time.Duration(c_time * 1e6))
			}
			p.T3 = time.Now().UnixNano()
			p.E3 = node.ntpClient.GetOffset()
			node.nc.Publish(fmt.Sprintf("%s.data", node.name), &p)
		})
	} else if *nodeType == "vehicle" {

		var state VehicleState
		var gps sensor_msgs.NavSatFix

		if *enableROS {
			n, err := goroslib.NewNode(goroslib.NodeConf{
				Name:          "goroslib_sub",
				MasterAddress: "localhost:11311",
			})
			if err != nil {
				panic(err)
			}
			defer n.Close()

			subState, err := goroslib.NewSubscriber(goroslib.SubscriberConf{
				Node:     n,
				Topic:    "state",
				Callback: func(msg *VehicleState) { state = *msg },
			})
			if err != nil {
				panic(err)
			}
			defer subState.Close()

			subGps, err := goroslib.NewSubscriber(goroslib.SubscriberConf{
				Node:     n,
				Topic:    "gps/filtered",
				Callback: func(msg *sensor_msgs.NavSatFix) { gps = *msg },
			})
			if err != nil {
				panic(err)
			}
			defer subGps.Close()
		}

		node = NewNode(*nodeName, natsClient, ntpClient, func(node *Node) {})
		node.nc.Subscribe("server.data", func(p *Packet) {
			p.Header.FrameID = node.name
			p.T4 = time.Now().UnixNano()
			p.E4 = node.ntpClient.GetOffset()
			p.X = state.X
			p.Y = state.Y
			p.Yaw = state.Yaw
			p.V = state.V
			p.Latitude = gps.Latitude
			p.Longitude = gps.Longitude
			p.Chk = Checksum(p.Data, p.Chk) // NOTE: After this, if chk == 0 then it's good. The message was not corrupted.
			p.Data = []byte{}               // NOTE: We empty it so all data isn't stored. Use for something else? Maybe time sync error?
			node.logs = append(node.logs, *p)
		})
	} else {
		log.Fatalf("Unsupported node type \"%s\".", *nodeType)
	}

	go node.ntpClient.QueryLoop(func(r ntp.Response) bool {
		val, _ := node.attr["alive"]
		return val != 0
	})

	node.run()
}
