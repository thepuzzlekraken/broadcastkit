package blackmagicdesign_test

import (
	"net"
	"reflect"
	"testing"
	"time"

	"puzzlekraken.com/broadcastkit/blackmagicdesign"
)

func TestSmoke(t *testing.T) {
	smoke := []byte(`PROTOCOL PREAMBLE:
Version: 2.8

VIDEOHUB DEVICE:
Device present: true
Model name: Blackmagic Smart Videohub 40 x 40
Friendly name: Blackmagic Smart Videohub 40 x 40
Unique ID: 7C2E0D038143
Video inputs: 40
Video processing units: 0
Video outputs: 40
Video monitoring outputs: 0
Serial ports: 0

INPUT LABELS:
0 INPUT 1
1 INPUT 2
2 INPUT 3
3 INPUT 4
4 INPUT 5
5 INPUT 6
6 INPUT 7
7 INPUT 8
8 INPUT 9
9 INPUT 10
10 INPUT 11
11 INPUT 12
12 INPUT 13
13 INPUT 14
14 INPUT 15
15 INPUT 16
16 INPUT 17
17 INPUT 18
18 INPUT 19
19 INPUT 20
20 INPUT 21
21 INPUT 22
22 INPUT 23
23 INPUT 24
24 INPUT 25
25 INPUT 26
26 INPUT 27
27 INPUT 28
28 INPUT 29
29 INPUT 30
30 INPUT 31
31 INPUT 32
32 INPUT 33
33 INPUT 34
34 INPUT 35
35 INPUT 36
36 INPUT 37
37 INPUT 38
38 INPUT 39
39 INPUT 40

OUTPUT LABELS:
0 OUTPUT 1
1 OUTPUT 2
2 OUTPUT 3
3 OUTPUT 4
4 OUTPUT 5
5 OUTPUT 6
6 OUTPUT 7
7 OUTPUT 8
8 OUTPUT 9
9 OUTPUT 10
10 OUTPUT 11
11 OUTPUT 12
12 OUTPUT 13
13 OUTPUT 14
14 OUTPUT 15
15 OUTPUT 16
16 OUTPUT 17
17 OUTPUT 18
18 OUTPUT 19
19 OUTPUT 20
20 OUTPUT 21
21 OUTPUT 22
22 OUTPUT 23
23 OUTPUT 24
24 OUTPUT 25
25 OUTPUT 26
26 OUTPUT 27
27 OUTPUT 28
28 OUTPUT 29
29 OUTPUT 30
30 OUTPUT 31
31 OUTPUT 32
32 OUTPUT 33
33 OUTPUT 34
34 OUTPUT 35
35 OUTPUT 36
36 OUTPUT 37
37 OUTPUT 38
38 OUTPUT 39
39 OUTPUT 40

VIDEO OUTPUT LOCKS:
0 U
1 U
2 U
3 U
4 U
5 U
6 U
7 U
8 U
9 U
10 U
11 L
12 O
13 U
14 U
15 U
16 U
17 U
18 U
19 U
20 U
21 U
22 U
23 U
24 U
25 U
26 U
27 U
28 U
29 U
30 U
31 U
32 U
33 U
34 U
35 U
36 U
37 U
38 U
39 U

VIDEO OUTPUT ROUTING:
0 0
1 1
2 2
3 3
4 4
5 5
6 6
7 7
8 8
9 9
10 10
11 11
12 12
13 13
14 3
15 15
16 16
17 17
18 18
19 19
20 20
21 21
22 22
23 23
24 24
25 25
26 26
27 27
28 28
29 29
30 30
31 31
32 32
33 33
34 34
35 35
36 36
37 37
38 38
39 39

CONFIGURATION:
Take Mode: false

END PRELUDE:

`)
	bacon := make([]blackmagicdesign.VideohubBlock, 0)
	bacon = append(bacon, &blackmagicdesign.ProtocolPreambleBlock{
		Version: struct {
			Major int
			Minor int
		}{
			Major: 2,
			Minor: 8,
		},
	})
	bacon = append(bacon, &blackmagicdesign.VideohubDeviceBlock{
		DevicePresent:          blackmagicdesign.DevicePresentTrue,
		ModelName:              "Blackmagic Smart Videohub 40 x 40",
		FriendlyName:           "Blackmagic Smart Videohub 40 x 40",
		UniqueID:               "7C2E0D038143",
		VideoInputs:            40,
		VideoProcessingUnits:   0,
		VideoOutputs:           40,
		VideoMonitoringOutputs: 0,
		SerialPorts:            0,
	})
	il := &blackmagicdesign.InputLabelsBlock{
		Labels: make(blackmagicdesign.Labels),
	}
	il.Labels[0] = "INPUT 1"
	il.Labels[1] = "INPUT 2"
	il.Labels[2] = "INPUT 3"
	il.Labels[3] = "INPUT 4"
	il.Labels[4] = "INPUT 5"
	il.Labels[5] = "INPUT 6"
	il.Labels[6] = "INPUT 7"
	il.Labels[7] = "INPUT 8"
	il.Labels[8] = "INPUT 9"
	il.Labels[9] = "INPUT 10"
	il.Labels[10] = "INPUT 11"
	il.Labels[11] = "INPUT 12"
	il.Labels[12] = "INPUT 13"
	il.Labels[13] = "INPUT 14"
	il.Labels[14] = "INPUT 15"
	il.Labels[15] = "INPUT 16"
	il.Labels[16] = "INPUT 17"
	il.Labels[17] = "INPUT 18"
	il.Labels[18] = "INPUT 19"
	il.Labels[19] = "INPUT 20"
	il.Labels[20] = "INPUT 21"
	il.Labels[21] = "INPUT 22"
	il.Labels[22] = "INPUT 23"
	il.Labels[23] = "INPUT 24"
	il.Labels[24] = "INPUT 25"
	il.Labels[25] = "INPUT 26"
	il.Labels[26] = "INPUT 27"
	il.Labels[27] = "INPUT 28"
	il.Labels[28] = "INPUT 29"
	il.Labels[29] = "INPUT 30"
	il.Labels[30] = "INPUT 31"
	il.Labels[31] = "INPUT 32"
	il.Labels[32] = "INPUT 33"
	il.Labels[33] = "INPUT 34"
	il.Labels[34] = "INPUT 35"
	il.Labels[35] = "INPUT 36"
	il.Labels[36] = "INPUT 37"
	il.Labels[37] = "INPUT 38"
	il.Labels[38] = "INPUT 39"
	il.Labels[39] = "INPUT 40"
	bacon = append(bacon, il)
	ol := &blackmagicdesign.OutputLabelsBlock{
		Labels: make(blackmagicdesign.Labels),
	}
	ol.Labels[0] = "OUTPUT 1"
	ol.Labels[1] = "OUTPUT 2"
	ol.Labels[2] = "OUTPUT 3"
	ol.Labels[3] = "OUTPUT 4"
	ol.Labels[4] = "OUTPUT 5"
	ol.Labels[5] = "OUTPUT 6"
	ol.Labels[6] = "OUTPUT 7"
	ol.Labels[7] = "OUTPUT 8"
	ol.Labels[8] = "OUTPUT 9"
	ol.Labels[9] = "OUTPUT 10"
	ol.Labels[10] = "OUTPUT 11"
	ol.Labels[11] = "OUTPUT 12"
	ol.Labels[12] = "OUTPUT 13"
	ol.Labels[13] = "OUTPUT 14"
	ol.Labels[14] = "OUTPUT 15"
	ol.Labels[15] = "OUTPUT 16"
	ol.Labels[16] = "OUTPUT 17"
	ol.Labels[17] = "OUTPUT 18"
	ol.Labels[18] = "OUTPUT 19"
	ol.Labels[19] = "OUTPUT 20"
	ol.Labels[20] = "OUTPUT 21"
	ol.Labels[21] = "OUTPUT 22"
	ol.Labels[22] = "OUTPUT 23"
	ol.Labels[23] = "OUTPUT 24"
	ol.Labels[24] = "OUTPUT 25"
	ol.Labels[25] = "OUTPUT 26"
	ol.Labels[26] = "OUTPUT 27"
	ol.Labels[27] = "OUTPUT 28"
	ol.Labels[28] = "OUTPUT 29"
	ol.Labels[29] = "OUTPUT 30"
	ol.Labels[30] = "OUTPUT 31"
	ol.Labels[31] = "OUTPUT 32"
	ol.Labels[32] = "OUTPUT 33"
	ol.Labels[33] = "OUTPUT 34"
	ol.Labels[34] = "OUTPUT 35"
	ol.Labels[35] = "OUTPUT 36"
	ol.Labels[36] = "OUTPUT 37"
	ol.Labels[37] = "OUTPUT 38"
	ol.Labels[38] = "OUTPUT 39"
	ol.Labels[39] = "OUTPUT 40"
	bacon = append(bacon, ol)
	vl := &blackmagicdesign.VideoOutputLocksBlock{
		Locks: make(blackmagicdesign.Locks),
	}
	vl.Locks[0] = blackmagicdesign.LockUnlocked
	vl.Locks[1] = blackmagicdesign.LockUnlocked
	vl.Locks[2] = blackmagicdesign.LockUnlocked
	vl.Locks[3] = blackmagicdesign.LockUnlocked
	vl.Locks[4] = blackmagicdesign.LockUnlocked
	vl.Locks[5] = blackmagicdesign.LockUnlocked
	vl.Locks[6] = blackmagicdesign.LockUnlocked
	vl.Locks[7] = blackmagicdesign.LockUnlocked
	vl.Locks[8] = blackmagicdesign.LockUnlocked
	vl.Locks[9] = blackmagicdesign.LockUnlocked
	vl.Locks[10] = blackmagicdesign.LockUnlocked
	vl.Locks[11] = blackmagicdesign.LockLocked
	vl.Locks[12] = blackmagicdesign.LockOwned
	vl.Locks[13] = blackmagicdesign.LockUnlocked
	vl.Locks[14] = blackmagicdesign.LockUnlocked
	vl.Locks[15] = blackmagicdesign.LockUnlocked
	vl.Locks[16] = blackmagicdesign.LockUnlocked
	vl.Locks[17] = blackmagicdesign.LockUnlocked
	vl.Locks[18] = blackmagicdesign.LockUnlocked
	vl.Locks[19] = blackmagicdesign.LockUnlocked
	vl.Locks[20] = blackmagicdesign.LockUnlocked
	vl.Locks[21] = blackmagicdesign.LockUnlocked
	vl.Locks[22] = blackmagicdesign.LockUnlocked
	vl.Locks[23] = blackmagicdesign.LockUnlocked
	vl.Locks[24] = blackmagicdesign.LockUnlocked
	vl.Locks[25] = blackmagicdesign.LockUnlocked
	vl.Locks[26] = blackmagicdesign.LockUnlocked
	vl.Locks[27] = blackmagicdesign.LockUnlocked
	vl.Locks[28] = blackmagicdesign.LockUnlocked
	vl.Locks[29] = blackmagicdesign.LockUnlocked
	vl.Locks[30] = blackmagicdesign.LockUnlocked
	vl.Locks[31] = blackmagicdesign.LockUnlocked
	vl.Locks[32] = blackmagicdesign.LockUnlocked
	vl.Locks[33] = blackmagicdesign.LockUnlocked
	vl.Locks[34] = blackmagicdesign.LockUnlocked
	vl.Locks[35] = blackmagicdesign.LockUnlocked
	vl.Locks[36] = blackmagicdesign.LockUnlocked
	vl.Locks[37] = blackmagicdesign.LockUnlocked
	vl.Locks[38] = blackmagicdesign.LockUnlocked
	vl.Locks[39] = blackmagicdesign.LockUnlocked
	bacon = append(bacon, vl)
	vor := &blackmagicdesign.VideoOutputRoutingBlock{
		Routing: make(blackmagicdesign.Routing),
	}
	vor.Routing[0] = 0
	vor.Routing[1] = 1
	vor.Routing[2] = 2
	vor.Routing[3] = 3
	vor.Routing[4] = 4
	vor.Routing[5] = 5
	vor.Routing[6] = 6
	vor.Routing[7] = 7
	vor.Routing[8] = 8
	vor.Routing[9] = 9
	vor.Routing[10] = 10
	vor.Routing[11] = 11
	vor.Routing[12] = 12
	vor.Routing[13] = 13
	vor.Routing[14] = 3
	vor.Routing[15] = 15
	vor.Routing[16] = 16
	vor.Routing[17] = 17
	vor.Routing[18] = 18
	vor.Routing[19] = 19
	vor.Routing[20] = 20
	vor.Routing[21] = 21
	vor.Routing[22] = 22
	vor.Routing[23] = 23
	vor.Routing[24] = 24
	vor.Routing[25] = 25
	vor.Routing[26] = 26
	vor.Routing[27] = 27
	vor.Routing[28] = 28
	vor.Routing[29] = 29
	vor.Routing[30] = 30
	vor.Routing[31] = 31
	vor.Routing[32] = 32
	vor.Routing[33] = 33
	vor.Routing[34] = 34
	vor.Routing[35] = 35
	vor.Routing[36] = 36
	vor.Routing[37] = 37
	vor.Routing[38] = 38
	vor.Routing[39] = 39
	bacon = append(bacon, vor)
	bacon = append(bacon, &blackmagicdesign.ConfigurationBlock{
		TakeMode: false,
	})
	bacon = append(bacon, &blackmagicdesign.EndPreludeBlock{})

	endpoint, err := net.Listen("tcp4", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	defer endpoint.Close()
	addr := endpoint.Addr().String()
	go func() {
		conn, err := endpoint.Accept()
		if err != nil {
			t.Fatal(err)
		}
		conn.Write(smoke)
		conn.Close()
	}()
	v, err := blackmagicdesign.VideohubDial(addr)
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	v.SetReadDeadline(time.Now().Add(time.Second * 5))
	for _, bite := range bacon {
		msg, err := v.Read()
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(msg, bite) {
			t.Fatalf("message %#v does not match %#v", msg, bite)
		}
	}
}
