package uselambda

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPayloadMustUnmarshal(t *testing.T) {
	payload := `{"name": "uselambda"}`

	type request struct {
		Name string `json:"name"`
	}

	req := Payload(payload).MustUnmarshal(new(request)).(*request)
	require.Equal(t, req.Name, "uselambda")
}
