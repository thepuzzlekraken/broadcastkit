package sony

import (
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

type CameraClient struct {
	Remote   netip.AddrPort
	Username string
	Password string
	Http     http.Client
	httpOnce sync.Once
}

const networkTimeout = 500 * time.Second

func (c *CameraClient) httpInit() {
	if c.Http.Timeout == 0 {
		c.Http.Timeout = networkTimeout
	}
	c.Http.Transport = &digest.Transport{
		Username:  c.Username,
		Password:  c.Password,
		Transport: c.Http.Transport,
	}
}

func (c *CameraClient) httpGet(ep endpoint, ps []Parameter) (*http.Response, error) {
	c.httpOnce.Do(c.httpInit)
	var host string
	if port := c.Remote.Port(); port == 0 || port == 80 {
		// Port 80 must be omitted from hostname for referer checks to succeed.
		host = c.Remote.Addr().String()
	} else {
		host = c.Remote.String()
	}

	u := url.URL{
		// Authentication handled by digest transport, don't not set username
		// and password here in the URL.
		Scheme: "http",
		Host:   host,
		// Trailing slash is required for the Referer header.
		Path: "/",
	}
	headers := make(http.Header)
	headers.Add("Referer", u.String())
	u.Path = "/command/" + string(ep) + ".cgi"

	v := make(url.Values)
	for _, p := range ps {
		v.Add(p.parameterKey(), p.parameterValue())
	}
	u.RawQuery = v.Encode()
	fmt.Printf("%s\n", u.String())
	return c.Http.Do(&http.Request{
		Method:     "GET",
		URL:        &u,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     headers,
		Body:       nil,
		Host:       host,
	})
}

func (c *CameraClient) set(ep endpoint, ps []Parameter) error {
	res, err := c.httpGet(ep, ps)
	if err != nil {
		return fmt.Errorf("parameter set error: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status: %s", res.Status)
	}
	return nil
}

func (c *CameraClient) inq(ep endpoint) ([]Parameter, error) {
	res, err := c.httpGet(inquiryEndpoint, []Parameter{inquiryParameter(ep)})
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
			p, err := createParameter(ep, key, v)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			parameters = append(parameters, p)
		}
	}
	return parameters, errors.Join(errs...)
}

const assignableEndpoint endpoint = "assignable"

type AssignableParameter interface {
	Parameter
	_assignableParameter()
}

func (c *CameraClient) SetAssignable(p ...AssignableParameter) error {
	return c.set(assignableEndpoint, castGeneric(p))
}
func (c *CameraClient) InqAssignable() ([]AssignableParameter, error) {
	gs, err := c.inq(assignableEndpoint)
	return castSpecific[AssignableParameter](gs), err
}

const ptzfEndpoint endpoint = "ptzf"

type PtzfParameter interface {
	Parameter
	_ptzfParameter()
}

func (c *CameraClient) SetPtzf(p ...PtzfParameter) error {
	return c.set(ptzfEndpoint, castGeneric(p))
}

func (c *CameraClient) InqPtzf() ([]PtzfParameter, error) {
	gs, err := c.inq(ptzfEndpoint)
	return castSpecific[PtzfParameter](gs), err
}
