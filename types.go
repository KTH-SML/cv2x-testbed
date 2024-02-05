package main

import (
	"github.com/bluenviron/goroslib/v2/pkg/msg"
	"github.com/bluenviron/goroslib/v2/pkg/msgs/std_msgs"
)

type GetRequest struct {
	Author string `json:"author"`
	Name   string `json:"name"`
}
type GetResponse struct {
	Data    int    `json:"data"`
	Success bool   `json:"success"`
	Reason  string `json:"reason"`
}
type GetMsg struct {
	Request  GetRequest  `json:"request"`
	Response GetResponse `json:"response"`
}

type SetRequest struct {
	Author string `json:"author"`
	Name   string `json:"name"`
	Data   int    `json:"data"`
}

type SetResponse struct {
	Success bool   `json:"success"`
	Reason  string `json:"reason"`
}

type SetMsg struct {
	Request  SetRequest  `json:"request"`
	Response SetResponse `json:"response"`
}

type header struct {
	Seq     int64  `json:"seq"`
	Stamp   int64  `json:"stamp"`
	FrameID string `json:"frame_id"`
}

type VehicleState struct {
	msg.Package  `ros:"svea_msgs"`
	Header       std_msgs.Header `rosname:"header"`
	ChildFrameId string          `rosname:"child_frame_id"`
	X            float64         `rosname:"x"`
	Y            float64         `rosname:"y"`
	Yaw          float64         `rosname:"yaw"`
	V            float32         `rosname:"v"`
	Covariance   [16]float64     `rosname:"covariance"`
}

type Packet struct {
	Header    header  `json:"header"`
	T1        int64   `json:"t1"`
	T2        int64   `json:"t2"`
	T3        int64   `json:"t3"`
	T4        int64   `json:"t4"`
	E1        int64   `json:"e1"`
	E2        int64   `json:"e2"`
	E3        int64   `json:"e3"`
	E4        int64   `json:"e4"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Yaw       float64 `json:"yaw"`
	V         float32 `json:"v"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Data      []byte  `json:"data"`
	Chk       int     `json:"chk"`
}
