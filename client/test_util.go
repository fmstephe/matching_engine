package client

import (
	"encoding/json"
)

func unmarshalResponse(bResp []byte) *response {
	resp := &response{}
	if err := json.Unmarshal(bResp, resp); err != nil {
		panic(err.Error())
	}
	return resp
}
