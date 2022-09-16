package external

import (
	"bytes"
	"net/http"

	juno "github.com/forbole/juno/v3/types"
	"github.com/rs/zerolog/log"
)

// HandleTx implements modules.TransactionModule
func (m *Module) HandleTx(tx *juno.Tx) error {
	log.Info().Str("txhash", tx.TxHash).Msg("✅✅✅✅ external HandleTx")

	txResponseJSON, err := m.cdc.MarshalJSON(tx.TxResponse)
	if err != nil {
		log.Err(err).Msg("Could no encode transaction response as json")
	}

	// log.Trace().Str("request", string(txResponseJSON)).Msg("Request >>>")

	requestBody := bytes.NewBuffer(txResponseJSON)

	log.Trace().Str("url", m.cfg.URL).Msg(">>>")
	response, err := http.Post(m.cfg.URL, "application/json", requestBody)
	if response != nil {
		log.Trace().Str("url", m.cfg.URL).Str("Response status", response.Status).Int("Response code", response.StatusCode).Msg("<<<")
		response.Body.Close()
	}
	if err != nil {
		log.Warn().Err(err).Msg("Failed to called external API")
	}

	return nil
}
