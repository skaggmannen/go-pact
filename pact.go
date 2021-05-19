package pact

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func DecodeV3(r io.Reader) (pact PactV3, err error) {
	err = json.NewDecoder(r).Decode(&pact)
	return
}

type PactV3 struct {
	Provider     Participant   `json:"provider"`
	Consumer     Participant   `json:"consumer"`
	Interactions []Interaction `json:"interactions"`
	Messages     []Message     `json:"messages"`
	MetaData     MetaData      `json:"metadata"`
}

func (p PactV3) Encode(w io.Writer) error {
	return json.NewEncoder(w).Encode(p)
}

func (p PactV3) VerifyHandler(t *testing.T, h http.Handler) {
	t.Run(fmt.Sprintf("between %s and %s", p.Consumer.Name, p.Provider.Name), func(t *testing.T) {
		for _, i := range p.Interactions {
			p.VerifyHandlerInteraction(t, h, i)
		}
	})
}

func (p PactV3) VerifyHandlerInteraction(t *testing.T, h http.Handler, i Interaction) {
	t.Run(i.Description, func(t *testing.T) {
		// TODO: Add support for states
		recorder := httptest.NewRecorder()
		h.ServeHTTP(recorder, i.Request.toHttpRequest())
		p.VerifyHandlerResponse(t, h, recorder.Result(), i.Response)
	})
}

func (p PactV3) VerifyHandlerResponse(t *testing.T, h http.Handler, rsp *http.Response, expect Response) {
	t.Run("status", func(t *testing.T) {
		assert.Equalf(t, expect.Status, rsp.StatusCode, "expected status %d but was %d", expect.Status, rsp.StatusCode)
	})
	t.Run("headers", func(t *testing.T) {
		// TODO: support matchers
		assert.Equal(t, expect.Headers.toHttp(), rsp.Header)
	})
	t.Run("body", func(t *testing.T) {
		// TODO: support matchers
		body, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}

		expectedBody, err := json.Marshal(expect.Body)
		if err != nil {
			panic(err)
		}

		if len(expectedBody) == 0 {
			return
		} else if len(body) == 0 {
			t.Fatal("response body was empty")
		}

		assert.JSONEq(t, string(expectedBody), string(body))
	})

}

type Participant struct {
	Name string `json:"name"`
}

type Interaction struct {
	Description    string                   `json:"description"`
	ProviderStates map[string]ProviderState `json:"providerStates"`
	Request        Request                  `json:"request"`
	Response       Response                 `json:"response"`
}

type ProviderState struct {
	Name   string                 `json:"name"`
	Params map[string]interface{} `json:"params"`
}

type Request struct {
	Method        string              `json:"method"`
	Path          string              `json:"path"`
	Query         map[string][]string `json:"query"`
	Headers       map[string]string   `json:"headers"`
	Body          interface{}         `json:"body"`
	MatchingRules MatchingRules       `json:"matchingRules"`
}

func (r Request) toHttpRequest() *http.Request {
	body, err := json.Marshal(r.Body)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest(r.Method, "https://dummy-url"+r.Path, bytes.NewReader(body))
	if err != nil {
		panic(err)
	}

	for k, v := range r.Headers {
		req.Header.Set(k, v)
	}

	return req
}

type Response struct {
	Status        int           `json:"status"`
	Headers       Header        `json:"headers"`
	Body          interface{}   `json:"body"`
	MatchingRules MatchingRules `json:"matchingRules"`
}

type Header map[string]string

func (h Header) toHttp() http.Header {
	httpHeader := make(http.Header)
	for k, v := range h {
		httpHeader.Set(k, v)
	}
	return httpHeader
}

type Message struct {
	Description    string                   `json:"description"`
	ProviderStates map[string]ProviderState `json:"providerStates"`
	MetaData       MetaData                 `json:"metaData"`
	Contents       interface{}              `json:"contents"`
	MatchingRules  MatchingRules            `json:"matchingRules"`
	Generators     Generators               `json:"generators"`
}

type MatchingRules struct {
	Path   MatchingRule            `json:"path"`
	Query  map[string]MatchingRule `json:"query"`
	Header map[string]MatchingRule `json:"header"`
	Body   map[string]MatchingRule `json:"body"`
}

type MatchingRule struct {
	Combine  string    `json:"combine"`
	Matchers []Matcher `json:"matchers"`
}

type Matcher struct {
	Match string `json:"match"`
	Regex string `json:"regex"`
	Min   int    `json:"min"`
	Max   int    `json:"max"`
	Value string `json:"value"`
}

type Generators struct {
	Body map[string]Generator `json:"body"`
}

type Generator struct {
	Type   string `json:"type"`
	Min    int    `json:"min"`
	Max    int    `json:"max"`
	Digits int    `json:"digits"`
	Size   int    `json:"size"`
	Regex  string `json:"regex"`
	Format string `json:"format"`
}

type MetaData map[string]interface{}
