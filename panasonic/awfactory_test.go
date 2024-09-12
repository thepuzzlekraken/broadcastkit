package panasonic

import (
	"fmt"
	"reflect"
	"testing"
)

func Test_RequestConsistency(t *testing.T) {
	for _, tt := range awRequestTable {
		t.Run(reflect.TypeOf(tt.new()).Elem().Name(), func(t *testing.T) {
			str1 := randomMatch(tt.sig)
			cmd := tt.new()
			cmd.unpackRequest(str1)
			str2 := cmd.packRequest()
			if str1 != str2 {
				t.Errorf("%v.packRequest() = %v, want %v", reflect.TypeOf(cmd), str2, str1)
			}
		})
	}
}

func Test_ResponseConsistency(t *testing.T) {
	for _, tt := range awResponseTable {
		t.Run(reflect.TypeOf(tt.new()).Elem().Name(), func(t *testing.T) {
			str1 := randomMatch(tt.sig)
			cmd := tt.new()
			cmd.unpackResponse(str1)
			str2 := cmd.packResponse()
			if str1 != str2 {
				t.Errorf("%v.packResponse() = %v, want %v", reflect.TypeOf(cmd), str2, str1)
			}
		})
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
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s (%s)", tt.name, tt.reqStr), func(t *testing.T) {
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
