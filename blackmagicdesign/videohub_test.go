package blackmagicdesign

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"testing"
)

func testRead(buf *bytes.Buffer, msg []VideohubBlock, t *testing.T) {
	v := VideohubSocket{
		Conn: buf,
	}
	for _, bite := range msg {
		msg, err := v.Read()
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(msg, bite) {
			t.Fatalf("message %#v does not match %#v", msg, bite)
		}
	}
	_, eof := v.Read()
	if !errors.Is(eof, io.EOF) {
		t.Fatalf("Videohub.Read() expected EOF, got %v", eof)
	}
}

func TestVideohubScoket_Read(t *testing.T) {
	testRead(bytes.NewBuffer(testSmoke), testBacon, t)
}

func TestVideohubSocket_Write(t *testing.T) {
	buf := new(bytes.Buffer)
	v := VideohubSocket{
		Conn: buf,
	}
	for _, bite := range testBacon {
		if err := v.Write(bite); err != nil {
			t.Fatal(err)
		}
	}
	testRead(buf, testBacon, t)
}

var testSmoke = []byte(`PROTOCOL PREAMBLE:
Version: 2.8

VIDEOHUB DEVICE:
Device present: true
Model name: Blackmagic Smart Videohub 40 x 40
Friendly name: My Videohub
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

var testBacon = []VideohubBlock{
	&ProtocolPreambleBlock{
		Version: struct {
			Major int
			Minor int
		}{
			Major: 2,
			Minor: 8,
		},
	},
	&VideohubDeviceBlock{
		DevicePresent:          DevicePresentTrue,
		ModelName:              "Blackmagic Smart Videohub 40 x 40",
		FriendlyName:           "My Videohub",
		UniqueID:               "7C2E0D038143",
		VideoInputs:            40,
		VideoProcessingUnits:   0,
		VideoOutputs:           40,
		VideoMonitoringOutputs: 0,
		SerialPorts:            0,
	},
	&InputLabelsBlock{
		Labels: Labels{
			0: "INPUT 1", 1: "INPUT 2", 2: "INPUT 3", 3: "INPUT 4",
			4: "INPUT 5", 5: "INPUT 6", 6: "INPUT 7", 7: "INPUT 8",
			8: "INPUT 9", 9: "INPUT 10", 10: "INPUT 11", 11: "INPUT 12",
			12: "INPUT 13", 13: "INPUT 14", 14: "INPUT 15", 15: "INPUT 16",
			16: "INPUT 17", 17: "INPUT 18", 18: "INPUT 19", 19: "INPUT 20",
			20: "INPUT 21", 21: "INPUT 22", 22: "INPUT 23", 23: "INPUT 24",
			24: "INPUT 25", 25: "INPUT 26", 26: "INPUT 27", 27: "INPUT 28",
			28: "INPUT 29", 29: "INPUT 30", 30: "INPUT 31", 31: "INPUT 32",
			32: "INPUT 33", 33: "INPUT 34", 34: "INPUT 35", 35: "INPUT 36",
			36: "INPUT 37", 37: "INPUT 38", 38: "INPUT 39", 39: "INPUT 40",
		},
	},
	&OutputLabelsBlock{
		Labels: Labels{
			0: "OUTPUT 1", 1: "OUTPUT 2", 2: "OUTPUT 3", 3: "OUTPUT 4",
			4: "OUTPUT 5", 5: "OUTPUT 6", 6: "OUTPUT 7", 7: "OUTPUT 8",
			8: "OUTPUT 9", 9: "OUTPUT 10", 10: "OUTPUT 11", 11: "OUTPUT 12",
			12: "OUTPUT 13", 13: "OUTPUT 14", 14: "OUTPUT 15", 15: "OUTPUT 16",
			16: "OUTPUT 17", 17: "OUTPUT 18", 18: "OUTPUT 19", 19: "OUTPUT 20",
			20: "OUTPUT 21", 21: "OUTPUT 22", 22: "OUTPUT 23", 23: "OUTPUT 24",
			24: "OUTPUT 25", 25: "OUTPUT 26", 26: "OUTPUT 27", 27: "OUTPUT 28",
			28: "OUTPUT 29", 29: "OUTPUT 30", 30: "OUTPUT 31", 31: "OUTPUT 32",
			32: "OUTPUT 33", 33: "OUTPUT 34", 34: "OUTPUT 35", 35: "OUTPUT 36",
			36: "OUTPUT 37", 37: "OUTPUT 38", 38: "OUTPUT 39", 39: "OUTPUT 40",
		},
	},
	&VideoOutputLocksBlock{
		Locks: Locks{
			0: LockUnlocked, 1: LockUnlocked, 2: LockUnlocked, 3: LockUnlocked,
			4: LockUnlocked, 5: LockUnlocked, 6: LockUnlocked, 7: LockUnlocked,
			8: LockUnlocked, 9: LockUnlocked, 10: LockUnlocked, 11: LockLocked,
			12: LockOwned, 13: LockUnlocked, 14: LockUnlocked, 15: LockUnlocked,
			16: LockUnlocked, 17: LockUnlocked, 18: LockUnlocked, 19: LockUnlocked,
			20: LockUnlocked, 21: LockUnlocked, 22: LockUnlocked, 23: LockUnlocked,
			24: LockUnlocked, 25: LockUnlocked, 26: LockUnlocked, 27: LockUnlocked,
			28: LockUnlocked, 29: LockUnlocked, 30: LockUnlocked, 31: LockUnlocked,
			32: LockUnlocked, 33: LockUnlocked, 34: LockUnlocked, 35: LockUnlocked,
			36: LockUnlocked, 37: LockUnlocked, 38: LockUnlocked, 39: LockUnlocked,
		},
	},
	&VideoOutputRoutingBlock{
		Routing: Routing{
			0: 0, 1: 1, 2: 2, 3: 3, 4: 4, 5: 5, 6: 6, 7: 7, 8: 8, 9: 9,
			10: 10, 11: 11, 12: 12, 13: 13, 14: 3, 15: 15, 16: 16, 17: 17, 18: 18, 19: 19,
			20: 20, 21: 21, 22: 22, 23: 23, 24: 24, 25: 25, 26: 26, 27: 27, 28: 28, 29: 29,
			30: 30, 31: 31, 32: 32, 33: 33, 34: 34, 35: 35, 36: 36, 37: 37, 38: 38, 39: 39,
		},
	},
	&ConfigurationBlock{
		TakeMode: false,
	},
	&EndPreludeBlock{},
}
