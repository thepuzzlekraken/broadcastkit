package panasonic

import (
	"fmt"
	"reflect"
	"slices"
	"testing"
)

func Test_RequestConsistency(t *testing.T) {
	for _, tt := range awRequestTable {
		name := reflect.TypeOf(tt.new()).Elem().Name()
		had := make([]string, 0, 4)
		for _, seed := range []int64{0, 746652860576448003, -2321909154513872682, -1} {
			str1 := generateMatch(tt.sig, seed)
			if slices.Contains(had, str1) {
				continue
			}
			had = append(had, str1)
			t.Run(name+" "+str1, func(t *testing.T) {
				cmd := tt.new()
				cmd.unpackRequest(str1)
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
		name := reflect.TypeOf(tt.new()).Elem().Name()
		had := make([]string, 0, 4)
		for _, seed := range []int64{0, -1, 746652860576448003, -2321909154513872682} {
			str1 := generateMatch(tt.sig, seed)
			if slices.Contains(had, str1) {
				continue
			}
			had = append(had, str1)
			t.Run(name+" "+str1, func(t *testing.T) {
				cmd := tt.new()
				cmd.unpackResponse(str1)
				str2 := cmd.packResponse()
				if str1 != str2 {
					t.Errorf("%v.packResponse() = %v, want %v", reflect.TypeOf(cmd), str2, str1)
				}
			})
		}
	}
}

func ExportedEqual(x, y any) bool {
	v1 := reflect.ValueOf(x).Elem()
	v2 := reflect.ValueOf(y).Elem()
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
			request:    &AWPower{Power: PowerOn},
			acceptable: true,
			response:   &AWPower{Power: PowerOn},
		},
		{
			name:       "Power Off",
			reqStr:     "#O0",
			request:    &AWPower{Power: PowerStandby},
			acceptable: true,
			response:   &AWPower{Power: PowerStandby},
		},
		{
			name:       "Power On Alternate",
			reqStr:     "#On",
			request:    &AWPower{Power: PowerOn},
			acceptable: true,
			response:   &AWPower{Power: PowerOn},
		},
		{
			name:       "Power Off Alternate",
			reqStr:     "#Of",
			request:    &AWPower{Power: PowerStandby},
			acceptable: true,
			response:   &AWPower{Power: PowerStandby},
		},
		{
			name:       "Power Query",
			reqStr:     "#O",
			request:    &AWPowerQuery{},
			acceptable: true,
			response:   &AWPower{},
		},
		{
			name:       "Install Position Desktop",
			reqStr:     "#INS0",
			request:    &AWInstall{Position: DesktopPosition},
			acceptable: true,
			response:   &AWInstall{Position: DesktopPosition},
		},
		{
			name:       "Install Postion Hanging",
			reqStr:     "#INS1",
			request:    &AWInstall{Position: HangingPosition},
			acceptable: true,
			response:   &AWInstall{Position: HangingPosition},
		},
		{
			name:       "Pan",
			reqStr:     "#P50",
			request:    &AWPan{Pan: 0},
			acceptable: true,
			response:   &AWPan{Pan: 0},
		},
		{
			name:       "Pan Up",
			reqStr:     "#P75",
			request:    &AWPan{Pan: 25},
			acceptable: true,
			response:   &AWPan{Pan: 25},
		},
		{
			name:       "Pan Down",
			reqStr:     "#P20",
			request:    &AWPan{Pan: -30},
			acceptable: true,
			response:   &AWPan{Pan: -30},
		},
		{
			name:       "Pan Invalid",
			reqStr:     "#P00",
			request:    &AWPan{Pan: -50},
			acceptable: false,
			response:   &AWPan{Pan: -50},
		},
		{
			name:       "Tilt",
			reqStr:     "#T50",
			request:    &AWTilt{Tilt: 0},
			acceptable: true,
			response:   &AWTilt{Tilt: 0},
		},
		{
			name:       "Tilt Up",
			reqStr:     "#T99",
			request:    &AWTilt{Tilt: 49},
			acceptable: true,
			response:   &AWTilt{Tilt: 49},
		},
		{
			name:       "Tilt Down",
			reqStr:     "#T01",
			request:    &AWTilt{Tilt: -49},
			acceptable: true,
			response:   &AWTilt{Tilt: -49},
		},
		{
			name:       "Preset Recall",
			reqStr:     "#R42",
			request:    &AWPresetRecall{Preset: 42},
			acceptable: true,
			response:   &AWPreset{Preset: 42},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s %s", tt.name, tt.reqStr), func(t *testing.T) {
			sig := tt.request.requestSignature()
			if !match(sig, tt.reqStr) {
				t.Errorf("match(%v,%v) = false, want true", sig, tt.reqStr)
			}
			acceptable := tt.request.Acceptable()
			if acceptable != tt.acceptable {
				t.Errorf("%v.Acceptable() = %v, want %v", tt.request, acceptable, tt.acceptable)
			}
			response := tt.request.Response()
			if !ExportedEqual(response, tt.response) {
				t.Errorf("Response() = %v, want %v", response, tt.response)
			}

			new := reflect.New(reflect.TypeOf(tt.request).Elem()).Interface().(AWRequest)
			new.unpackRequest(tt.reqStr)
			if !ExportedEqual(new, tt.request) {
				t.Errorf("unpackRequest(%v) -> %v, want %v", tt.reqStr, new, tt.request)
			}
			pack := new.packRequest()
			if pack != tt.reqStr {
				t.Errorf("%v.packRequest() = %v, want %v", tt.request, pack, tt.reqStr)
			}
		})
	}
}
