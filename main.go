package main

import (
	"net/netip"
	"time"

	"puzzlekraken.com/broadcastkit/sony"
)

func main() {
	/*
		cam := panasonic.CameraClient{
			Remote: netip.AddrPortFrom(netip.MustParseAddr("192.168.5.53"), 80),
		}
		b, err := cam.AWBatch()
		if err != nil {
			panic(err)
		}
		for _, c := range b {
			fmt.Printf("%#v\n", c)
		}*/
	cam := sony.CameraClient{
		Remote:   netip.AddrPortFrom(netip.MustParseAddr("192.168.5.52"), 80),
		Username: "admin",
		Password: "Hummer1234",
	}
	for {
		err := cam.SetPtzf(sony.AbsolutePanTiltParam{
			Pan:   -30 * sony.SteppedPositionByDegree,
			Tilt:  -10 * sony.SteppedPositionByDegree,
			Speed: 50,
		})
		if err != nil {
			panic(err)
		}
		time.Sleep(time.Second * 3)
		err = cam.SetPtzf(sony.AbsolutePanTiltParam{
			Pan:   +30 * sony.SteppedPositionByDegree,
			Tilt:  +10 * sony.SteppedPositionByDegree,
			Speed: 50,
		})
		if err != nil {
			panic(err)
		}
		time.Sleep(time.Second * 3)
	}
}
