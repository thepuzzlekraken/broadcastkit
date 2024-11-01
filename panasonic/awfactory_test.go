package panasonic

import (
	"fmt"
	"reflect"
	"slices"
	"testing"
)

func Test_RequestConsistency(t *testing.T) {
	for _, tt := range awRequestTable {
		name := reflect.TypeOf(tt.new()).Name()
		had := make([]string, 0, 4)
		for _, seed := range []int64{0, 746652860576448003, -2321909154513872682, -1} {
			str1 := generateMatch(tt.sig, seed)
			if slices.Contains(had, str1) {
				continue
			}
			had = append(had, str1)
			t.Run(name+" "+str1, func(t *testing.T) {
				cmd := tt.new().unpackRequest(str1)
				str2 := cmd.packRequest()
				if str1 != str2 {
					t.Errorf("%v.packRequest() = %v, want %v", reflect.TypeOf(cmd), str2, str1)
				}
			})
		}
	}
}

func Test_ResponseConsistency(t *testing.T) {
	for _, tt := range awResponseTable {
		name := reflect.TypeOf(tt.new()).Name()
		had := make([]string, 0, 4)
		for _, seed := range []int64{0, -1, 746652860576448003, -2321909154513872682} {
			str1 := generateMatch(tt.sig, seed)
			if slices.Contains(had, str1) {
				continue
			}
			had = append(had, str1)
			t.Run(name+" "+str1, func(t *testing.T) {
				cmd := tt.new().unpackResponse(str1)
				str2 := cmd.packResponse()
				if str1 != str2 {
					t.Errorf("%v.packResponse() = %v, want %v", reflect.TypeOf(cmd), str2, str1)
				}
			})
		}
	}
}

func Test_RequestConflicts(t *testing.T) {
	seed := int64(48613)
	for _, tt := range awRequestTable {
		t.Run(reflect.TypeOf(tt.new()).Name(), func(t *testing.T) {
			cmd1 := tt.new()
			ref1 := reflect.TypeOf(cmd1)
			sig1 := cmd1.requestSignature()
			if sig1 != tt.sig {
				t.Errorf("%v.requestSignature() = %v, want %v", cmd1, cmd1.requestSignature(), tt.sig)
			}
			str := generateMatch(tt.sig, seed)
			cmd2 := newRequest(str)
			ref2 := reflect.TypeOf(cmd2)
			sig2 := cmd2.requestSignature()
			if sig2 != tt.sig {
				t.Errorf("%v.requestSignature() = %v, want %v", cmd2, cmd2.requestSignature(), tt.sig)
			}
			if ref1 != ref2 {
				t.Errorf("conflict in tables for %s, %v != %v", str, ref1, ref2)
			}
		})
		seed++
	}
}

func Test_ResponseConflicts(t *testing.T) {
	seed := int64(48613)
	for _, tt := range awResponseTable {
		t.Run(reflect.TypeOf(tt.new()).Name(), func(t *testing.T) {
			cmd1 := tt.new()
			ref1 := reflect.TypeOf(cmd1)
			sig1 := cmd1.responseSignature()
			if sig1 != tt.sig {
				t.Errorf("%v.responseSignature() = %v, want %v", cmd1, cmd1.responseSignature(), tt.sig)
			}
			str := generateMatch(tt.sig, seed)
			cmd2 := newResponse(str)
			ref2 := reflect.TypeOf(cmd2)
			sig2 := cmd2.responseSignature()
			if sig2 != tt.sig {
				t.Errorf("%v.responseSignature() = %v, want %v", cmd2, cmd2.responseSignature(), tt.sig)
			}
			if ref1 != ref2 {
				t.Errorf("conflict in tables for %s, %v != %v", str, ref1, ref2)
			}
		})
		seed++
	}
}

func ExportedEqual(x, y any) bool {
	v1 := reflect.ValueOf(x)
	v2 := reflect.ValueOf(y)
	t := v1.Type()
	if v1.Kind() != reflect.Struct {
		return reflect.DeepEqual(x, y)
	}
	if v1.Type() != v2.Type() {
		return false
	}
	for i := 0; i < v1.NumField(); i++ {
		if !t.Field(i).IsExported() {
			continue
		}
		if !reflect.DeepEqual(v1.Field(i).Interface(), v2.Field(i).Interface()) {
			return false
		}
	}
	return true
}

func TestAWRequest(t *testing.T) {
	tests := []struct {
		name       string
		reqStr     string
		request    AWRequest
		acceptable bool
		response   AWResponse
	}{
		{
			name:       "Power On",
			reqStr:     "#O1",
			request:    AWPower{Power: PowerOn},
			acceptable: true,
			response:   AWPower{Power: PowerOn},
		},
		{
			name:       "Power Off",
			reqStr:     "#O0",
			request:    AWPower{Power: PowerStandby},
			acceptable: true,
			response:   AWPower{Power: PowerStandby},
		},
		{
			name:       "Power On Alternate",
			reqStr:     "#On",
			request:    AWPower{Power: PowerOn},
			acceptable: true,
			response:   AWPower{Power: PowerOn},
		},
		{
			name:       "Power Off Alternate",
			reqStr:     "#Of",
			request:    AWPower{Power: PowerStandby},
			acceptable: true,
			response:   AWPower{Power: PowerStandby},
		},
		{
			name:       "Power Query",
			reqStr:     "#O",
			request:    AWPowerQuery{},
			acceptable: true,
			response:   AWPower{},
		},
		{
			name:       "Install Position Desktop",
			reqStr:     "#INS0",
			request:    AWInstall{Position: DesktopPosition},
			acceptable: true,
			response:   AWInstall{Position: DesktopPosition},
		},
		{
			name:       "Install Postion Hanging",
			reqStr:     "#INS1",
			request:    AWInstall{Position: HangingPosition},
			acceptable: true,
			response:   AWInstall{Position: HangingPosition},
		},
		{
			name:       "Pan",
			reqStr:     "#P50",
			request:    AWPan{Pan: 0},
			acceptable: true,
			response:   AWPan{Pan: 0},
		},
		{
			name:       "Pan Up",
			reqStr:     "#P75",
			request:    AWPan{Pan: 25},
			acceptable: true,
			response:   AWPan{Pan: 25},
		},
		{
			name:       "Pan Down",
			reqStr:     "#P20",
			request:    AWPan{Pan: -30},
			acceptable: true,
			response:   AWPan{Pan: -30},
		},
		{
			name:       "Pan Invalid",
			reqStr:     "#P00",
			request:    AWPan{Pan: -50},
			acceptable: false,
			response:   AWPan{Pan: -50},
		},
		{
			name:       "Tilt",
			reqStr:     "#T50",
			request:    AWTilt{Tilt: 0},
			acceptable: true,
			response:   AWTilt{Tilt: 0},
		},
		{
			name:       "Tilt Up",
			reqStr:     "#T99",
			request:    AWTilt{Tilt: 49},
			acceptable: true,
			response:   AWTilt{Tilt: 49},
		},
		{
			name:       "Tilt Down",
			reqStr:     "#T01",
			request:    AWTilt{Tilt: -49},
			acceptable: true,
			response:   AWTilt{Tilt: -49},
		},
		{
			name:       "Preset Recall",
			reqStr:     "#R42",
			request:    AWPresetRecall{Preset: 42},
			acceptable: true,
			response:   AWPreset{Preset: 42},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s %s", tt.name, tt.reqStr), func(t *testing.T) {
			sig := tt.request.requestSignature()
			if !match(sig, tt.reqStr) {
				t.Errorf("match(%v,%v) = false, want true", sig, tt.reqStr)
			}
			new := newRequest(tt.reqStr)
			if !ExportedEqual(new, tt.request) {
				t.Errorf("unpackRequest(%v) -> %v, want %v", tt.reqStr, new, tt.request)
			}
			acceptable := tt.request.Acceptable()
			if acceptable != tt.acceptable {
				t.Errorf("%v.Acceptable() = %v, want %v", tt.request, acceptable, tt.acceptable)
			}
			response := tt.request.Response()
			if !ExportedEqual(response, tt.response) {
				t.Errorf("Response() = %v, want %v", response, tt.response)
			}
			pack := new.packRequest()
			if pack != tt.reqStr {
				t.Errorf("%v.packRequest() = %v, want %v", tt.request, pack, tt.reqStr)
			}
		})
	}
}

func TestAWResponse(t *testing.T) {
	tests := []struct {
		name     string
		resStr   string
		response AWResponse
	}{
		{
			name:     "Power On",
			resStr:   "p1",
			response: AWPower{Power: PowerOn},
		},
		{
			name:     "Power Off",
			resStr:   "p0",
			response: AWPower{Power: PowerStandby},
		},
		{
			name:     "Install Position Desktop",
			resStr:   "iNS0",
			response: AWInstall{Position: DesktopPosition},
		},
		{
			name:     "Install Postion Hanging",
			resStr:   "iNS1",
			response: AWInstall{Position: HangingPosition},
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s %s", tt.name, tt.resStr), func(t *testing.T) {
			sig := tt.response.responseSignature()
			if !match(sig, tt.resStr) {
				t.Errorf("match(%v,%v) = false, want true", sig, tt.resStr)
			}
			new := newResponse(tt.resStr)
			if !ExportedEqual(new, tt.response) {
				t.Errorf("newResponse(%v) -> %v, want %v", tt.resStr, new, tt.response)
			}
			pack := new.packResponse()
			if pack != tt.resStr {
				t.Errorf("%v.packResponse() = %v, want %v", tt.response, pack, tt.resStr)
			}
		})
	}
}
