package main

import (
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

type Node struct {
	name      string
	attr      map[string]int
	main      func(*Node)
	logs      []Packet
	nc        *nats.EncodedConn
	ntpClient *NTPClient
}

func NewNode(name string, nc *nats.EncodedConn, ntpClient *NTPClient, main func(*Node)) *Node {

	node := &Node{
		name:      name,
		attr:      map[string]int{"rate": 1, "paused": 1, "alive": 1},
		main:      main,
		nc:        nc,
		ntpClient: ntpClient,
	}

	// SETUP own setters and getters
	nc.Subscribe(fmt.Sprintf("%s.get", name), node.get_srv_cb)
	nc.Subscribe(fmt.Sprintf("%s.get.log", name), node.get_srv_log_cb)
	nc.Subscribe(fmt.Sprintf("%s.set", name), node.set_srv_cb)
	nc.Subscribe(fmt.Sprintf("%s.set.log", name), node.set_srv_log_cb)
	return node
}

func (n *Node) get_srv_cb(subj, reply string, msg GetRequest) {
	if val, ok := n.attr[msg.Name]; ok {
		resp := &GetResponse{Success: true, Data: val}
		err := n.nc.Publish(reply, resp)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Getting", msg.Name, "by", msg.Author)

	} else {
		resp := &GetResponse{Success: false, Reason: fmt.Sprintf("Trying to write non-existing field \"%s\"", msg.Name)}
		err := n.nc.Publish(reply, resp)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (n *Node) get_srv_log_cb(subj, reply string, _ SetRequest) {
	timeNow := time.Now().Format("060102_1504")
	fileName := fmt.Sprintf("%s.csv", timeNow)
	save(n.logs, fileName)
	n.nc.Publish(reply, n.logs)
}

func (n *Node) set_srv_cb(subj, reply string, msg SetRequest) {
	if _, ok := n.attr[msg.Name]; ok {
		n.attr[msg.Name] = msg.Data
		n.nc.Publish(reply, &SetResponse{Success: true})
		fmt.Println("Setting", msg.Name, "to", msg.Data, "by", msg.Author)
	} else {
		resp := &SetResponse{Success: false, Reason: fmt.Sprintf("Trying to write non-existing field \"%s\"", msg.Name)}
		n.nc.Publish(reply, resp)
	}
}

func (n *Node) set_srv_log_cb(subj, reply string, msg []Packet) {
	n.logs = msg
	n.nc.Publish(reply, &SetResponse{Success: true})
}

func (n *Node) kill(remote_names ...string) error {
	for _, remote_name := range remote_names {
		_, err := n.remote_set(remote_name, "alive", 0)
		if err != nil {
			return fmt.Errorf("%w to \"%s\"", err, remote_name)
		}
	}
	return nil
}

func (n *Node) pause(remote_names ...string) error {
	for _, remote_name := range remote_names {
		_, err := n.remote_set(remote_name, "paused", 1)
		if err != nil {
			return fmt.Errorf("%w to \"%s\"", err, remote_name)
		}
	}
	return nil
}

func (n *Node) unpause(remote_names ...string) error {
	for _, remote_name := range remote_names {
		_, err := n.remote_set(remote_name, "paused", 0)
		if err != nil {
			return fmt.Errorf("%w to \"%s\"", err, remote_name)
		}
	}
	return nil
}

func (n *Node) remote_get(remote_name string, attr_name string) (GetResponse, error) {
	req := &GetRequest{Author: n.name, Name: attr_name}
	var resp GetResponse
	err := n.nc.Request(fmt.Sprintf("%s.get", remote_name), req, &resp, time.Second)
	if err != nil {
		return GetResponse{}, err
	}
	return resp, nil
}

func (n *Node) remote_get_log(remote_name string) ([]Packet, error) {
	req := &GetRequest{Author: n.name}
	var resp []Packet
	err := n.nc.Request(fmt.Sprintf("%s.get.log", remote_name), req, &resp, time.Second)
	if err != nil {
		return []Packet{}, err
	}
	return resp, nil
}

func (n *Node) remote_set(remote_name string, attr_name string, value int) (SetResponse, error) {
	req := &SetRequest{Author: n.name, Name: attr_name, Data: value}
	var resp SetResponse
	err := n.nc.Request(fmt.Sprintf("%s.set", remote_name), req, &resp, time.Second)
	if err != nil {
		return SetResponse{}, err
	}
	return resp, nil
}

func (n *Node) remote_set_log(remote_name string, value []Packet) (SetResponse, error) {
	var resp SetResponse
	err := n.nc.Request(fmt.Sprintf("%s.set.log", remote_name), value, &resp, time.Second)
	if err != nil {
		return SetResponse{}, err
	}
	return resp, nil
}

func (n *Node) isAlive() bool {
	val, _ := n.attr["alive"]
	return val != 0
}

func (n *Node) isPaused() bool {
	val, _ := n.attr["paused"]
	return val != 0
}

func (n *Node) run() {
	for {
		if val, _ := n.attr["alive"]; val == 0 {
			break
		}
		if val, _ := n.attr["paused"]; val == 0 {
			n.main(n)
		}
		rate, _ := n.attr["rate"]
		time.Sleep(time.Duration(1e9 / rate))
	}
}
