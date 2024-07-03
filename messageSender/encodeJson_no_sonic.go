//go:build !sonic || !avx || !(windows || linux || darwin) || !amd64

package messageSender

import "bytes"
import "encoding/json"

func encodeJson(data any, buf *bytes.Buffer) error {
	return json.NewEncoder(buf).Encode(data)
}
