package runtime

import (
	"errors"
	"time"
)

type ToolApprovalPolicy string

type ToolSafetyClass string

const (
	ToolApprovalAlwaysRequired ToolApprovalPolicy = "always_required"
	ToolApprovalNotRequired    ToolApprovalPolicy = "not_required"

	ToolSafetyNoSideEffectInternal ToolSafetyClass = "no_side_effect_internal"
)

type ToolArguments struct {
	Timezone string
}

type ToolDefinition struct {
	Name           string
	ApprovalPolicy ToolApprovalPolicy
	SafetyClass    ToolSafetyClass
}

func CurrentTimeToolDefinition() ToolDefinition {
	return ToolDefinition{Name: "runtime.get_current_time", ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyNoSideEffectInternal}
}

func (d ToolDefinition) Execute(arguments ToolArguments) (map[string]any, error) {
	if d.Name != "runtime.get_current_time" {
		return nil, errors.New("tool is not supported")
	}
	if arguments.Timezone != "UTC" {
		return nil, errors.New("timezone must be UTC")
	}
	return map[string]any{"iso_time": time.Now().UTC().Format(time.RFC3339Nano), "timezone": "UTC", "source": "runtime"}, nil
}

func (d ToolDefinition) NormalizeArguments(arguments map[string]any) (ToolArguments, error) {
	if d.Name != "runtime.get_current_time" {
		return ToolArguments{}, errors.New("tool is not supported")
	}
	for key := range arguments {
		if key != "timezone" {
			return ToolArguments{}, errors.New("tool argument is not supported")
		}
	}
	value, ok := arguments["timezone"]
	if !ok || value == nil {
		return ToolArguments{Timezone: "UTC"}, nil
	}
	timezone, ok := value.(string)
	if !ok || timezone != "UTC" {
		return ToolArguments{}, errors.New("timezone must be UTC")
	}
	return ToolArguments{Timezone: "UTC"}, nil
}
