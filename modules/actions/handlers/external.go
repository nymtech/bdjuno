package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/forbole/bdjuno/v3/modules/actions/types"

	"github.com/rs/zerolog/log"
)

func ExternalRequestHandler(ctx *types.Context, payload *types.Payload) (interface{}, error) {
	log.Debug().Str("address", payload.GetAddress()).
		Int64("height", payload.Input.Height).
		Msg("executing external action")

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	requestBody := bytes.NewBuffer(payloadJSON)

	// TODO: parse the URL from the payload and sanitise
	url := "http://localhost:3001/api/v1/test"

	log.Trace().Str("url", url).Msg(">>>")
	response, err := http.Post(url, "application/json", requestBody)
	log.Trace().Str("url", url).Str("Response status", response.Status).Int("Response code", response.StatusCode).Msg("<<<")
	if err != nil {
		response.Body.Close()
	}

	return types.Coin{Denom: "foo", Amount: "0"}, nil
}
