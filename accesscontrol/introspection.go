package accesscontrol

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/hashicorp/hcl/v2"

	"github.com/avenga/couper/cache"
	"github.com/avenga/couper/config"
	"github.com/avenga/couper/config/request"
	"github.com/avenga/couper/eval"
	"github.com/avenga/couper/oauth2"
)

type lock struct {
	mu sync.Mutex
}

// Introspector represents a token introspector.
type Introspector struct {
	authenticator *oauth2.ClientAuthenticator
	conf          *config.Introspection
	locks         sync.Map
	memStore      *cache.MemoryStore
	transport     http.RoundTripper
}

// NewIntrospector creates a new token introspector.
func NewIntrospector(evalCtx *hcl.EvalContext, conf *config.Introspection, transport http.RoundTripper, memStore *cache.MemoryStore) (*Introspector, error) {
	authenticator, err := oauth2.NewClientAuthenticator(evalCtx, conf.EndpointAuthMethod, "endpoint_auth_method", conf.ClientID, conf.ClientSecret, "", conf.JWTSigningProfile)
	if err != nil {
		return nil, err
	}
	return &Introspector{
		authenticator: authenticator,
		conf:          conf,
		memStore:      memStore,
		transport:     transport,
	}, nil
}

// IntrospectionResponse represents the response body to a token introspection request.
type IntrospectionResponse map[string]interface{}

// Active returns whether the token is active.
func (ir IntrospectionResponse) Active() bool {
	active, _ := ir["active"].(bool)
	return active
}

func (ir IntrospectionResponse) exp() int64 {
	exp, _ := ir["exp"].(int64)
	return exp
}

// Introspect retrieves introspection data for the given token using either cached or fresh information.
func (i *Introspector) Introspect(ctx context.Context, token string, exp, nbf int64) (IntrospectionResponse, error) {
	var (
		introspectionData IntrospectionResponse
		key               string
	)

	if i.conf.TTLSeconds > 0 {
		// lock per token
		entry, _ := i.locks.LoadOrStore(token, &lock{})
		l := entry.(*lock)
		l.mu.Lock()
		defer func() {
			i.locks.Delete(token)
			l.mu.Unlock()
		}()

		key = "ir:" + token
		cachedIntrospectionBytes, _ := i.memStore.Get(key).([]byte)
		if cachedIntrospectionBytes != nil {
			// cached introspection response is always JSON
			_ = json.Unmarshal(cachedIntrospectionBytes, &introspectionData)

			// return cached introspection data
			return introspectionData, nil
		}
	}

	req, _ := http.NewRequest("POST", i.conf.Endpoint, nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	outCtx, cancel := context.WithCancel(context.WithValue(ctx, request.RoundTripName, "introspection"))
	defer cancel()

	formParams := &url.Values{}
	formParams.Add("token", token)

	err := i.authenticator.Authenticate(formParams, req)
	if err != nil {
		return nil, err
	}

	eval.SetBody(req, []byte(formParams.Encode()))

	req = req.WithContext(outCtx)

	response, err := i.transport.RoundTrip(req)
	if err != nil {
		return nil, fmt.Errorf("introspection response: %s", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("introspection response status code %d", response.StatusCode)
	}

	resBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("introspection response cannot be read: %s", err)
	}

	err = json.Unmarshal(resBytes, &introspectionData)
	if err != nil {
		return nil, fmt.Errorf("introspection response is not JSON: %s", err)
	}

	if i.conf.TTLSeconds <= 0 {
		return introspectionData, nil
	}

	if exp == 0 {
		if isdExp := introspectionData.exp(); isdExp > 0 {
			exp = isdExp
		}
	}

	ttl := i.conf.TTLSeconds

	if exp > 0 {
		now := time.Now().Unix()
		maxTTL := exp - now
		if !introspectionData.Active() && (nbf <= 0 || now > nbf) {
			// nbf is unknown (token has never been inactive before being active)
			// or nbf lies in the past (token has become active after having been inactive):
			// token will not become active again, so we can store the response until token expires anyway
			ttl = maxTTL
		} else if ttl > maxTTL {
			ttl = maxTTL
		}
	}
	// cache introspection data
	i.memStore.Set(key, resBytes, ttl)

	return introspectionData, nil
}
