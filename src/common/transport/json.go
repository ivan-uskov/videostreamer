package transport

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
)

func MarshalJson(v any) []byte {
	res, err := json.Marshal(v)
	if err != nil {
		log.WithError(err).Fatal("marshal")
	}

	return res
}

func UnmarshalJson(data []byte, v any) {
	err := json.Unmarshal(data, v)
	if err != nil {
		log.WithError(err).Fatal("unmarshal")
	}
}
