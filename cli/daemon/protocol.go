package daemon

import (
	"github.com/DoniLite/Mogoly/sync"
)


// Helper to decode payload from sync.Message
func DecodePayload(msg *sync.Message, target interface{}) error {
	return msg.DecodePayload(target)
}

func NewSuccessMessage(reqID string, actionType sync.Action_Type, data interface{}) (*sync.Message, error) {
	msg, err := sync.NewMessage(actionType, data, nil)
	if err != nil {
		return nil, err
	}
	msg.RequestID = reqID
	return msg, nil
}

func NewErrorMessage(reqID string, actionType sync.Action_Type, errMsg string) *sync.Message {
	// sync.NewErrorMessage creates a message with type ERROR.
	// We might want to preserve the action type or at least the reqID.
	msg := sync.NewErrorMessage(errMsg, "")
	msg.RequestID = reqID
	// We could set the action type to the original request type if we want,
	// but sync.NewErrorMessage sets it to ERROR.
	// Let's stick with ERROR type for errors as per sync package design.
	return msg
}
