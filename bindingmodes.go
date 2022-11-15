package i3

import "encoding/json"

// GetBindingModes returns the names of all currently configured binding modes.
//
// GetBindingModes is supported in i3 ≥ v4.13 (2016-11-08).
func GetBindingModes() ([]string, error) {
	reply, err := roundTrip(messageTypeGetBindingModes, nil)
	if err != nil {
		return nil, err
	}

	var bm []string
	err = json.Unmarshal(reply.Payload, &bm)
	return bm, err
}

// BindingState indicates which binding mode is currently active.
type BindingState struct {
	Name string `json:"name"`
}

// GetBindingState returns the currently active binding mode.
//
// GetBindingState is supported in i3 ≥ 4.19 (2020-11-15).
func GetBindingState() (BindingState, error) {
	reply, err := roundTrip(messageTypeGetBindingState, nil)
	if err != nil {
		return BindingState{}, err
	}

	var bm BindingState
	err = json.Unmarshal(reply.Payload, &bm)
	return bm, err
}
