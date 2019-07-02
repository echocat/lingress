package context

import (
	"errors"
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
	ErrIllegalStage = errors.New("illegal stage")

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

	nameToStage = func(in map[Stage]string) map[string]Stage {
		out := make(map[string]Stage)
		for stage, name := range in {
			out[name] = stage
		}
		return out
	}(stageToName)
)

func ParseStage(plain string) (Stage, error) {
	if stage, ok := nameToStage[plain]; ok {
		return stage, nil
	} else {
		return 0, ErrIllegalStage
	}
}

func (instance Stage) String() string {
	if name, ok := stageToName[instance]; ok {
		return name
	} else {
		return fmt.Sprintf("unknown-stage-%d", instance)
	}
}
