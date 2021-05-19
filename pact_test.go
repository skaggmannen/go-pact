package pact

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeV3(t *testing.T) {
	t.Run("spec example", func(t *testing.T) {
		p, err := DecodeV3(mustOpenFile("./testdata/pact-v3-example.json"))
		require.NoError(t, err)
		require.NotZero(t, p)
	})
}

func TestPactV3_VerifyHandler(t *testing.T) {
	p, err := DecodeV3(mustOpenFile("./testdata/pact-v3-example.json"))
	require.NoError(t, err)

	p.VerifyHandler(t, http.HandlerFunc(func(rsp http.ResponseWriter, req *http.Request) {
		rsp.Header().Set("Content-Type", "application/json; charset=UTF-8")
		rsp.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(rsp).Encode([]map[string]interface{}{
			{
				"dob":       "07/19/2016",
				"id":        8958464620,
				"name":      "Rogger the Dogger",
				"timestamp": "2016-07-19T12:14:39",
			},
			{
				"dob":       "07/19/2016",
				"id":        4143398442,
				"name":      "Cat in the Hat",
				"timestamp": "2016-07-19T12:14:39",
			},
		})
	}))
}

func mustOpenFile(path string) *os.File {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	return f
}
