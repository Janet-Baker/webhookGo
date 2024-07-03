//go:build sonic && avx && (linux || windows || darwin) && amd64

package messageSender

import (
	"bytes"
	"github.com/bytedance/sonic"
)

var api = sonic.ConfigDefault

func encodeJson(data any, buf *bytes.Buffer) error {
	return api.NewEncoder(buf).Encode(data)
}
