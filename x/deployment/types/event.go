package types

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ovrclk/akash/sdkutil"
)

const (
	evActionDeploymentCreate = "deployment-create"
	evActionDeploymentUpdate = "deployment-update"
	evActionDeploymentClose  = "deployment-close"
	evActionGroupClose       = "group-close"
	evOwnerKey               = "owner"
	evDSeqKey                = "dseq"
	evGSeqKey                = "gseq"
)

// EventDeploymentCreate struct
type EventDeploymentCreate struct {
	sdkutil.BaseModuleEvent `json:",inline"`
	ID                      DeploymentID `json:"id"`
}

func NewEventDeploymentCreate(id DeploymentID) EventDeploymentCreate {
	return EventDeploymentCreate{
		BaseModuleEvent: sdkutil.BaseModuleEvent{
			Module: ModuleName,
			Action: evActionDeploymentCreate,
		},
		ID: id,
	}
}

// ToSDKEvent method creates new sdk event for EventDeploymentCreate struct
func (ev EventDeploymentCreate) ToSDKEvent() sdk.Event {
	return sdk.NewEvent(sdkutil.EventTypeMessage,
		append([]sdk.Attribute{
			sdk.NewAttribute(sdk.AttributeKeyModule, ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, evActionDeploymentCreate),
		}, DeploymentIDEVAttributes(ev.ID)...)...,
	)
}

// EventDeploymentUpdate struct
type EventDeploymentUpdate struct {
	sdkutil.BaseModuleEvent `json:",inline"`
	ID                      DeploymentID `json:"id"`
	Version                 []byte       `json:"version,omitempty"` // TODO
}

func NewEventDeploymentUpdate(id DeploymentID) EventDeploymentUpdate {
	return EventDeploymentUpdate{
		BaseModuleEvent: sdkutil.BaseModuleEvent{
			Module: ModuleName,
			Action: evActionDeploymentUpdate,
		},
		ID: id,
	}
}

// ToSDKEvent method creates new sdk event for EventDeploymentUpdate struct
func (ev EventDeploymentUpdate) ToSDKEvent() sdk.Event {
	return sdk.NewEvent(sdkutil.EventTypeMessage,
		append([]sdk.Attribute{
			sdk.NewAttribute(sdk.AttributeKeyModule, ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, evActionDeploymentUpdate),
		}, DeploymentIDEVAttributes(ev.ID)...)...,
	)
}

// EventDeploymentClose struct
type EventDeploymentClose struct {
	sdkutil.BaseModuleEvent `json:",inline"`
	ID                      DeploymentID `json:"id"`
}

func NewEventDeploymentClose(id DeploymentID) EventDeploymentClose {
	return EventDeploymentClose{
		BaseModuleEvent: sdkutil.BaseModuleEvent{
			Module: ModuleName,
			Action: evActionDeploymentClose,
		},
		ID: id,
	}
}

// ToSDKEvent method creates new sdk event for EventDeploymentClose struct
func (ev EventDeploymentClose) ToSDKEvent() sdk.Event {
	return sdk.NewEvent(sdkutil.EventTypeMessage,
		append([]sdk.Attribute{
			sdk.NewAttribute(sdk.AttributeKeyModule, ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, evActionDeploymentClose),
		}, DeploymentIDEVAttributes(ev.ID)...)...,
	)
}

// DeploymentIDEVAttributes returns event attribues for given DeploymentID
func DeploymentIDEVAttributes(id DeploymentID) []sdk.Attribute {
	return []sdk.Attribute{
		sdk.NewAttribute(evOwnerKey, id.Owner.String()),
		sdk.NewAttribute(evDSeqKey, strconv.FormatUint(id.DSeq, 10)),
	}
}

// ParseEVDeploymentID returns deploymentID details for given event attributes
func ParseEVDeploymentID(attrs []sdk.Attribute) (DeploymentID, error) {
	owner, err := sdkutil.GetAccAddress(attrs, evOwnerKey)
	if err != nil {
		return DeploymentID{}, err
	}
	dseq, err := sdkutil.GetUint64(attrs, evDSeqKey)
	if err != nil {
		return DeploymentID{}, err
	}

	return DeploymentID{
		Owner: owner,
		DSeq:  dseq,
	}, nil
}

// EventGroupClose provides SDK event to signal group termination
type EventGroupClose struct {
	sdkutil.BaseModuleEvent `json:",inline"`
	ID                      GroupID `json:"id"`
}

func NewEventGroupClose(id GroupID) EventGroupClose {
	return EventGroupClose{
		BaseModuleEvent: sdkutil.BaseModuleEvent{
			Module: ModuleName,
			Action: evActionGroupClose,
		},
		ID: id,
	}
}

// ToSDKEvent produces the SDK notification for Event
func (ev EventGroupClose) ToSDKEvent() sdk.Event {
	return sdk.NewEvent(sdkutil.EventTypeMessage,
		append([]sdk.Attribute{
			sdk.NewAttribute(sdk.AttributeKeyModule, ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, evActionGroupClose),
		}, GroupIDEVAttributes(ev.ID)...)...,
	)
}

// GroupIDEVAttributes returns event attribues for given GroupID
func GroupIDEVAttributes(id GroupID) []sdk.Attribute {
	return append(DeploymentIDEVAttributes(id.DeploymentID()),
		sdk.NewAttribute(evGSeqKey, strconv.FormatUint(uint64(id.GSeq), 10)))
}

// ParseEVGroupID returns GroupID details for given event attributes
func ParseEVGroupID(attrs []sdk.Attribute) (GroupID, error) {
	did, err := ParseEVDeploymentID(attrs)
	if err != nil {
		return GroupID{}, err
	}

	gseq, err := sdkutil.GetUint64(attrs, evGSeqKey)
	if err != nil {
		return GroupID{}, err
	}

	return GroupID{
		Owner: did.Owner,
		DSeq:  did.DSeq,
		GSeq:  uint32(gseq),
	}, nil
}

// ParseEvent parses event and returns details of event and error if occurred
// TODO: Enable returning actual events.
func ParseEvent(ev sdkutil.Event) (sdkutil.ModuleEvent, error) {
	if ev.Type != sdkutil.EventTypeMessage {
		return nil, sdkutil.ErrUnknownType
	}
	if ev.Module != ModuleName {
		return nil, sdkutil.ErrUnknownModule
	}
	switch ev.Action {
	case evActionDeploymentCreate:
		did, err := ParseEVDeploymentID(ev.Attributes)
		if err != nil {
			return nil, err
		}
		return NewEventDeploymentCreate(did), nil
	case evActionDeploymentUpdate:
		did, err := ParseEVDeploymentID(ev.Attributes)
		if err != nil {
			return nil, err
		}
		return NewEventDeploymentUpdate(did), nil
	case evActionDeploymentClose:
		did, err := ParseEVDeploymentID(ev.Attributes)
		if err != nil {
			return nil, err
		}
		return NewEventDeploymentClose(did), nil
	case evActionGroupClose:
		gid, err := ParseEVGroupID(ev.Attributes)
		if err != nil {
			return nil, err
		}
		return NewEventGroupClose(gid), nil
	default:
		return nil, sdkutil.ErrUnknownAction
	}
}
