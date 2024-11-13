package sony

import (
	"strconv"
	"time"
)

const inquiryEndpoint Endpoint = "inquiry"

type inqParam Endpoint

func (inqParam) parameterKey() string {
	return "inq"
}
func (ep inqParam) parameterValue() string {
	return string(ep) // cut leading slash and trailing .cgi
}
func (inqParam) parameterParse(s string) (Parameter, error) {
	return inqParam(s), nil
}
func (inqParam) Valid() bool {
	return true
}

const subscribeEndpoint Endpoint = "subscribe"

type inqjsonParam Endpoint

func (inqjsonParam) parameterKey() string {
	return "inqjson"
}
func (ep inqjsonParam) parameterValue() string {
	return string(ep) // cut leading slash and trailing .cgi
}
func (inqjsonParam) parameterParse(s string) (Parameter, error) {
	return inqjsonParam(s), nil
}
func (inqjsonParam) Valid() bool {
	return true
}

type subscriptionDurationParam time.Duration

func (subscriptionDurationParam) parameterKey() string {
	return "SubscriptionDuration"
}
func (d subscriptionDurationParam) parameterValue() string {
	return itoa(int(d / subscriptionDurationParam(time.Second)))
}
func (subscriptionDurationParam) parameterParse(s string) (Parameter, error) {
	i, err := atoi(s)
	if err != nil {
		return nil, err
	}
	return subscriptionDurationParam(time.Duration(i) * time.Second), nil
}
func (subscriptionDurationParam) Valid() bool {
	return true
}

const pullinqueryEndpoint Endpoint = "pullinquiry"
const unsubscribeEndpoint Endpoint = "unsubscribe"

type SubscriptionIdParam string

func (SubscriptionIdParam) parameterKey() string {
	return "SubscriptionId"
}
func (ep SubscriptionIdParam) parameterValue() string {
	return string(ep)
}
func (SubscriptionIdParam) parameterParse(s string) (Parameter, error) {
	return SubscriptionIdParam(s), nil
}
func (SubscriptionIdParam) Valid() bool {
	return true
}

type cacheKillParam struct{}

func (cacheKillParam) parameterKey() string {
	return "_"
}
func (cacheKillParam) parameterValue() string {
	return strconv.FormatInt(time.Now().UnixMilli(), 10)
}
func (cacheKillParam) parameterParse(s string) (Parameter, error) {
	return cacheKillParam{}, nil
}
func (cacheKillParam) Valid() bool {
	return true
}
