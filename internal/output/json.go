package output

import (
	"encoding/json"
	"io"
)

// JSONRenderer renders data as JSON.
type JSONRenderer struct{}

func NewJSONRenderer() *JSONRenderer {
	return &JSONRenderer{}
}

func (r *JSONRenderer) RenderNetworks(w io.Writer, nets []NetworkRow) error {
	return encodeJSON(w, nets)
}

func (r *JSONRenderer) RenderInterfaceInfo(w io.Writer, info map[string]string) error {
	return encodeJSON(w, info)
}

func (r *JSONRenderer) RenderSpeedResult(w io.Writer, result map[string]string) error {
	return encodeJSON(w, result)
}

func (r *JSONRenderer) RenderSignal(w io.Writer, data SignalData) error {
	return encodeJSON(w, data)
}

func (r *JSONRenderer) RenderChannelAnalysis(w io.Writer, channels []ChannelInfo) error {
	return encodeJSON(w, channels)
}

func (r *JSONRenderer) RenderHealthReport(w io.Writer, data *HealthReportData) error {
	return encodeJSON(w, data)
}

func encodeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
