package sony

import (
	"reflect"
	"strings"
	"testing"
)

var implements map[Endpoint]reflect.Type

func init() {
	implements = make(map[Endpoint]reflect.Type)
	implements[AssignableEndpoint] = reflect.TypeFor[AssignableParameter]()
	implements[PtzfEndpoint] = reflect.TypeFor[PtzfParameter]()
	implements[PresetpositionEndpoint] = reflect.TypeFor[PresetpositionParameter]()
}

func TestParamNames(t *testing.T) {
	for key, new := range parameterTable {
		t.Run(key, func(t *testing.T) {
			p := new()
			if pkey := p.parameterKey(); key != pkey {
				t.Errorf("parameter %s incorrectly keyed %s", key, pkey)
			}
			if tkey := reflect.TypeOf(p).Name(); key+"Param" != tkey {
				t.Errorf("parameter %s has type name %s, expects <Key>Param", key, tkey)
			}
		})
	}
}

func TestInterfaceNames(t *testing.T) {
	for ep, iface := range implements {
		t.Run(string(ep), func(t *testing.T) {
			tname := iface.Name()
			expected := strings.ToTitle(string(ep[0])) + string(ep[1:]) + "Parameter"
			if tname != expected {
				t.Errorf("interface for %s has type name %s, expects <Key>Parameter", ep, tname)
			}
		})
	}
}
