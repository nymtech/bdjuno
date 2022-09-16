package handlers

import (
	dbtypes "github.com/forbole/bdjuno/v3/database/types"
	"github.com/forbole/bdjuno/v3/modules/actions/types"
	"github.com/rs/zerolog/log"
)

type MessagesPaging struct {
	Limit      uint64 `json:"limit"`
	Offset     uint64 `json:"offset"`
	TotalCount int    `json:"total_count"`
}

type MessagesParams struct {
	Limit       uint64
	Offset      uint64
	StartHeight int64
	EndHeight   int64
	Address     string
}

func ValidateMessageParams(ctx *types.Context, payload *types.Payload, message string) (*MessagesParams, error) {
	log.Debug().Str("address", payload.GetAddress()).
		Time("executed_at_start", payload.Input.ExecutedAtStart).
		Time("executed_at_end", payload.Input.ExecutedAtEnd).
		Uint64("offset", payload.Input.Offset).
		Uint64("limit", payload.Input.Limit).
		Str("location", "before").
		Msg(message)

	limit := uint64(10)
	offset := uint64(0)

	executedAtStart := payload.Input.ExecutedAtStart
	executedAtEnd := payload.Input.ExecutedAtEnd

	startHeight := int64(-1)
	endHeight := int64(-1)

	valueStart, err := ctx.Db.GetBlockHeightTime(executedAtStart)
	if err != nil {
		valueFirst, err := ctx.Db.GetFirstBlockTime()
		if err != nil {
			return nil, err
		} else {
			startHeight = valueFirst.Height
		}
	} else {
		startHeight = valueStart.Height
	}

	valueEnd, err := ctx.Db.GetBlockHeightTime(executedAtEnd)
	if err != nil {
		valueLast, err := ctx.Db.GetLastBlockTime()
		if err != nil {
			return nil, err
		} else {
			startHeight = valueLast.Height
		}
	} else {
		endHeight = valueEnd.Height
	}

	address := payload.Input.Address

	if payload.Input.Limit > 10 || payload.Input.Limit < 1 {
		limit = uint64(10)
	} else {
		limit = payload.Input.Limit
	}
	if payload.Input.Offset < 1 {
		offset = uint64(0)
	} else {
		offset = payload.Input.Offset
	}

	log.Debug().Str("address", address).
		Int64("start_height", startHeight).
		Int64("end_height", endHeight).
		Uint64("offset", offset).
		Uint64("limit", limit).
		Str("location", "after").
		Msg(message)

	output := MessagesParams{
		Limit:       limit,
		Offset:      offset,
		StartHeight: startHeight,
		EndHeight:   endHeight,
		Address:     address,
	}

	return &output, nil
}

func MessagesHandler(ctx *types.Context, payload *types.Payload) (interface{}, error) {

	params, err := ValidateMessageParams(ctx, payload, "message handler")
	if err != nil {
		return nil, err
	}

	rows, err := ctx.Db.GetMessages(params.Address, params.StartHeight, params.EndHeight, params.Offset, params.Limit)
	if err != nil {
		return nil, err
	}

	rowsWithValue := make([]dbtypes.MessageWithValueRow, len(rows))

	for i, row := range rows {
		newRow, err := dbtypes.NewMessageWithValueRow(row)
		if err != nil {
			return nil, err
		}
		rowsWithValue[i] = *newRow
	}

	return rowsWithValue, err
}

func MessagesCountHandler(ctx *types.Context, payload *types.Payload) (interface{}, error) {
	params, err := ValidateMessageParams(ctx, payload, "message count handler")
	if err != nil {
		return nil, err
	}

	count, err := ctx.Db.GetMessagesCount(params.Address, params.StartHeight, params.EndHeight)
	if err != nil {
		return nil, err
	}

	output := MessagesPaging{
		Limit:      params.Limit,
		Offset:     params.Offset,
		TotalCount: count,
	}

	return output, err
}
