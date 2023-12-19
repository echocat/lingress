package context

import (
	"fmt"
)

type Stage uint8

const (
	StageUnknown                Stage = 0
	StageCreated                Stage = 1
	StageEvaluateClientRequest  Stage = 2
	StagePrepareUpstreamRequest Stage = 3
	StageSendRequestToUpstream  Stage = 4
	StagePrepareClientResponse  Stage = 5
	StageSendResponseToClient   Stage = 6
	StageDone                   Stage = 7
)

var (
	stageToName = map[Stage]string{
		StageUnknown:                "unknown",
		StageCreated:                "created",
		StageEvaluateClientRequest:  "evaluateClientRequest",
		StagePrepareUpstreamRequest: "prepareUpstreamRequest",
		StageSendRequestToUpstream:  "sendRequestToUpstream",
		StagePrepareClientResponse:  "prepareClientResponse",
		StageSendResponseToClient:   "sendResponseToClient",
		StageDone:                   "done",
	}
)

func (this Stage) String() string {
	if name, ok := stageToName[this]; ok {
		return name
	} else {
		return fmt.Sprintf("unknown-stage-%d", this)
	}
}
