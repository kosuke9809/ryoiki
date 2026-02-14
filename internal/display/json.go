package display

import (
	"encoding/json"
	"io"
)

// PrintJSON outputs data as formatted JSON
func PrintJSON(w io.Writer, data interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}
