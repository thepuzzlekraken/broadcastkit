package sony

type UnknownParameter struct {
	key   string
	value string
}

func (p UnknownParameter) parameterKey() string {
	return p.key
}
func (p UnknownParameter) parameterValue() string {
	return p.value
}
func (p UnknownParameter) parameterParse(val string) (Parameter, error) {
	return UnknownParameter{
		key:   p.key,
		value: val,
	}, nil
}
func (p UnknownParameter) Valid() bool {
	return true
}

type Parameter interface {
	Valid() bool
	parameterKey() string
	parameterValue() string
	parameterParse(string) (Parameter, error)
}

var parameterTable = make(map[string]func() Parameter)

func registerParameter(new func() Parameter) {
	key := new().parameterKey()
	parameterTable[key] = new
}

func createParameter(key string, val string) (Parameter, error) {
	new, ok := parameterTable[key]
	if !ok {
		return UnknownParameter{
			key:   key,
			value: val,
		}, nil
	}
	p, err := new().parameterParse(val)
	if err != nil {
		return nil, err
	}
	return p, nil
}
