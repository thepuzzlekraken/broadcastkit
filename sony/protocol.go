package sony

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"net/url"
	"sync"
	"time"

	"github.com/icholy/digest"
)

type Endpoint string

type CameraClient struct {
	Remote   netip.AddrPort
	Username string
	Password string

	Http     http.Client
	longHttp http.Client
	referer  string
	httpOnce sync.Once
}

const networkTimeout = 5 * time.Second
const pullPeriod = 60 * time.Second

func (c *CameraClient) httpInit() {
	c.longHttp = c.Http

	if c.Http.Timeout == 0 {
		c.Http.Timeout = networkTimeout
	}
	c.longHttp.Timeout = pullPeriod + networkTimeout

	c.Http.Transport = &digest.Transport{
		Username:  c.Username,
		Password:  c.Password,
		Transport: c.Http.Transport,
	}
	c.longHttp.Transport = &digest.Transport{
		Username:  c.Username,
		Password:  c.Password,
		Transport: c.longHttp.Transport,
	}
}

func (c *CameraClient) httpReq(ctx context.Context, ep Endpoint, ps ...Parameter) *http.Request {
	var u url.URL

	u.Scheme = "http"

	if port := c.Remote.Port(); port == 0 || port == 80 {
		// Port 80 must be omitted from hostname for referer checks to succeed.
		u.Host = c.Remote.Addr().String()
	} else {
		u.Host = c.Remote.String()
	}

	u.Path = "/" // Terminating slash is required for the Referer header
	referer := u.String()
	if ep != "" {
		u.Path = "/command/" + string(ep) + ".cgi"
	}

	v := make(url.Values, len(ps))
	for _, p := range ps {
		v.Add(p.parameterKey(), p.parameterValue())
	}
	u.RawQuery = v.Encode()

	r, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		panic(err)
	}
	r.Header.Set("Referer", referer)

	return r
}

func (c *CameraClient) Set(ep Endpoint, ps []Parameter) error {
	c.httpOnce.Do(c.httpInit)
	res, err := c.Http.Do(c.httpReq(context.Background(), ep, ps...))

	if err != nil {
		return fmt.Errorf("parameter set error: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status: %s", res.Status)
	}

	return nil
}

func (c *CameraClient) Inq(ep Endpoint) ([]Parameter, error) {
	c.httpOnce.Do(c.httpInit)
	res, err := c.Http.Do(c.httpReq(context.Background(), inquiryEndpoint, inqParam(ep)))

	if err != nil {
		return nil, fmt.Errorf("parameter inquery error: %w", err)
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("parameter read error: %w", err)
	}

	values, err := url.ParseQuery(string(data))
	if err != nil {
		return nil, fmt.Errorf("parameter parse error: %w", err)
	}

	parameters := make([]Parameter, 0, len(values))
	var errs []error
	for key, list := range values {
		for _, v := range list {
			p, err := createParameter(key, v)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			parameters = append(parameters, p)
		}
	}

	return parameters, errors.Join(errs...)
}

func (c *CameraClient) Subscribe(ep ...Endpoint) (SubscriptionIdParam, error) {
	c.httpOnce.Do(c.httpInit)

	params := make([]Parameter, 0, len(ep)+1)
	params = append(params, subscriptionDurationParam(pullPeriod))
	for _, e := range ep {
		params = append(params, inqjsonParam(e))
	}

	res, err := c.Http.Do(c.httpReq(context.Background(), subscribeEndpoint, params...))
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected http status: %s", res.Status)
	}

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("subscription body read error: %w", err)
	}
	data := struct {
		Id string `json:"subscription_id"`
		// TODO(zsh): Do we need to care about other fields? Seems like not...
	}{}
	if err := json.Unmarshal(raw, &data); err != nil {
		return "", fmt.Errorf("subscription body parse error: %w", err)
	}

	return SubscriptionIdParam(data.Id), nil
}

func (c *CameraClient) doPull(ctx context.Context, id SubscriptionIdParam) ([]byte, error) {
	res, err := c.longHttp.Do(c.httpReq(ctx, pullinqueryEndpoint, id, cacheKillParam{}))
	if err != nil {
		return nil, fmt.Errorf("pull error: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNoContent {
		return nil, nil
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected http status: %s", res.Status)
	}
	return io.ReadAll(res.Body)
}

func parsePull(raw []byte) ([]Parameter, error) {
	data := make(map[string]map[string]string)
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("pull parse error: %w", err)
	}
	var parameters []Parameter
	var errs []error
	for _, list := range data {
		for key, val := range list {
			p, err := createParameter(key, val)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			parameters = append(parameters, p)
		}
	}
	return parameters, errors.Join(errs...)
}

func (c *CameraClient) PullInq(ctx context.Context, id SubscriptionIdParam) ([]Parameter, error) {
	c.httpOnce.Do(c.httpInit)
	for {
		data, err := c.doPull(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("pull error: %w", err)
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if data == nil {
			continue
		}
		return parsePull(data)
	}
}

func (c *CameraClient) Unsubscribe(id SubscriptionIdParam) error {
	c.httpOnce.Do(c.httpInit)
	res, err := c.Http.Do(c.httpReq(context.Background(), unsubscribeEndpoint, id))
	if err != nil {
		return fmt.Errorf("unsubscribe error: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status: %s", res.Status)
	}
	return nil
}

const AssignableEndpoint Endpoint = "assignable"

type AssignableParameter interface {
	Parameter
	_assignableParameter()
}

func (c *CameraClient) SetAssignable(p ...AssignableParameter) error {
	return c.Set(AssignableEndpoint, castGeneric(p))
}
func (c *CameraClient) InqAssignable() ([]AssignableParameter, error) {
	gs, err := c.Inq(AssignableEndpoint)
	return castSpecific[AssignableParameter](gs), err
}

const CameraoperationEndpoint = "cameraoperation"

type CameraoperationParameter interface {
	Parameter
	_cameraoperationParameter()
}

func (c *CameraClient) SetCameraoperation(p ...CameraoperationParameter) error {
	return c.Set(CameraoperationEndpoint, castGeneric(p))
}
func (c *CameraClient) InqCameraoperation() ([]CameraoperationParameter, error) {
	gs, err := c.Inq(CameraoperationEndpoint)
	return castSpecific[CameraoperationParameter](gs), err
}

const PtzfEndpoint Endpoint = "ptzf"

type PtzfParameter interface {
	Parameter
	_ptzfParameter()
}

func (c *CameraClient) SetPtzf(p ...PtzfParameter) error {
	return c.Set(PtzfEndpoint, castGeneric(p))
}

func (c *CameraClient) InqPtzf() ([]PtzfParameter, error) {
	gs, err := c.Inq(PtzfEndpoint)
	return castSpecific[PtzfParameter](gs), err
}

const PresetpositionEndpoint Endpoint = "presetposition"

type PresetpositionParameter interface {
	Parameter
	_presetpositionParameter()
}

func (c *CameraClient) SetPresetposition(p ...PresetpositionParameter) error {
	return c.Set(PresetpositionEndpoint, castGeneric(p))
}
func (c *CameraClient) InqPresetposition() ([]PresetpositionParameter, error) {
	gs, err := c.Inq(PresetpositionEndpoint)
	return castSpecific[PresetpositionParameter](gs), err
}
