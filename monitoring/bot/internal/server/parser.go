package server

import "encoding/json"

func ParseAlert(data []byte) (AlertmanagerPayload, error) {
	var payload AlertmanagerPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return AlertmanagerPayload{}, err
	}
	return payload, nil
}
