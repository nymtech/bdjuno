package types

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestWasmTestSuite(t *testing.T) {
	suite.Run(t, new(WasmTestSuite))
}

type WasmTestSuite struct {
	suite.Suite
}

func (suite *WasmTestSuite) TestBigDipperTypes_CorrectlyParseWasmContractExecuteMessageType_Empty() {
	var wasmContractExecuteMessage string

	// nil returns nil type
	messageType := GetWasmExecuteContractMessageType([]byte(wasmContractExecuteMessage))
	suite.Require().Equal("", messageType)

	// empty string returns nil type
	wasmContractExecuteMessage = ""
	messageType = GetWasmExecuteContractMessageType([]byte(wasmContractExecuteMessage))
	suite.Require().Equal("", messageType)
}

func (suite *WasmTestSuite) TestBigDipperTypes_CorrectlyParseWasmContractExecuteMessageType_ParseErrorIsNil() {
	wasmContractExecuteMessage := "invalid JSON should not return a message type"
	messageType := GetWasmExecuteContractMessageType([]byte(wasmContractExecuteMessage))
	suite.Require().Equal("", messageType)
}

func (suite *WasmTestSuite) TestBigDipperTypes_CorrectlyParseWasmContractExecuteMessageType_ArrayIsNil() {
	wasmContractExecuteMessage := `[{ "foo": 1 }, { "bar": 2 }]`
	messageType := GetWasmExecuteContractMessageType([]byte(wasmContractExecuteMessage))
	suite.Require().Equal("", messageType)
}

func (suite *WasmTestSuite) TestBigDipperTypes_CorrectlyParseWasmContractExecuteMessageType_SingleValueIsNil() {
	wasmContractExecuteMessage := `"this should not return a message type"`
	messageType := GetWasmExecuteContractMessageType([]byte(wasmContractExecuteMessage))
	suite.Require().Equal("", messageType)
}

func (suite *WasmTestSuite) TestBigDipperTypes_CorrectlyParseWasmContractExecuteMessageType_SingleKeyParses() {
	wasmContractExecuteMessage := `{
    "foo": {
        "arg1": 1,
        "arg2": true,
        "arg3": {
            "arg4": "4"
        }
    }
}`
	messageType := GetWasmExecuteContractMessageType([]byte(wasmContractExecuteMessage))
	suite.Require().Equal("foo", messageType)
}

func (suite *WasmTestSuite) TestBigDipperTypes_CorrectlyParseWasmContractExecuteMessageType_TwoKeysParses() {
	wasmContractExecuteMessage := `{
    "foo": "1",
    "bar": "2"
}`
	messageType := GetWasmExecuteContractMessageType([]byte(wasmContractExecuteMessage))
	suite.Require().Equal("foo,bar", messageType)
}
