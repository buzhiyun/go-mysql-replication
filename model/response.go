package model

import (
	"sync"
)

var mqRespondPool = sync.Pool{
	New: func() interface{} {
		return new(MQResponse)
	},
}

type MQResponse struct {
	Topic     string      `json:"-"`
	Action    string      `json:"action"`
	Timestamp uint32      `json:"timestamp"`
	Schema    string      `json:"schema"`
	Table     string      `json:"table"`
	Raw       interface{} `json:"raw,omitempty"`
	Values    interface{} `json:"values"`
	OldValues interface{} `json:"oldvalues,omitempty"`
	ByteArray []byte      `json:"-"`
}
