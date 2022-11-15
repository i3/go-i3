package i3

import "encoding/json"

// IncludedConfig represents a single file that i3 has read, either because the
// file is the main config file, or because the file is included.
//
// IncludedConfig is supported in i3 ≥ v4.20 (2021-10-19).
type IncludedConfig struct {
	Path                     string `json:"path"`
	RawContents              string `json:"raw_contents"`
	VariableReplacedContents string `json:"variable_replaced_contents"`
}

// Config contains details about the configuration file.
//
// See https://i3wm.org/docs/ipc.html#_config_reply for more details.
type Config struct {
	Config string `json:"config"`

	// The IncludedConfigs field was added in i3 v4.20 (2021-10-19).
	IncludedConfigs []IncludedConfig `json:"included_configs"`
}

// GetConfig returns i3’s in-memory copy of the configuration file contents.
//
// GetConfig is supported in i3 ≥ v4.14 (2017-09-04).
func GetConfig() (Config, error) {
	reply, err := roundTrip(messageTypeGetConfig, nil)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	err = json.Unmarshal(reply.Payload, &cfg)
	return cfg, err
}
