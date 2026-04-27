package claude

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ctru0009/ccswap/internal/config"
)

// MergeEnv replaces only the 6 target env keys in the settings map with values
// from the given provider, preserving all other keys. It calls config.ExpandProvider()
// first for env var interpolation. If auth_token is empty after expansion, it returns
// an error. The input settings map is not mutated; a new copy is returned.
func MergeEnv(settings map[string]json.RawMessage, provider config.Provider) (map[string]json.RawMessage, error) {
	expanded, err := config.ExpandProvider(provider)
	if err != nil {
		return nil, fmt.Errorf("expanding provider: %w", err)
	}

	result := make(map[string]json.RawMessage, len(settings))
	for k, v := range settings {
		result[k] = v
	}

	var env map[string]string
	if envRaw, ok := result["env"]; ok {
		if err := json.Unmarshal(envRaw, &env); err != nil {
			return nil, fmt.Errorf("parsing existing env block: %w", err)
		}
	}
	if env == nil {
		env = make(map[string]string)
	}

	env["ANTHROPIC_AUTH_TOKEN"] = expanded.AuthToken
	env["ANTHROPIC_BASE_URL"] = expanded.BaseURL
	env["ANTHROPIC_DEFAULT_SONNET_MODEL"] = expanded.Models.Sonnet
	env["ANTHROPIC_DEFAULT_OPUS_MODEL"] = expanded.Models.Opus
	env["ANTHROPIC_DEFAULT_HAIKU_MODEL"] = expanded.Models.Haiku
	env["API_TIMEOUT_MS"] = strconv.Itoa(expanded.TimeoutMs)

	envBytes, err := json.Marshal(env)
	if err != nil {
		return nil, fmt.Errorf("marshaling env block: %w", err)
	}
	result["env"] = json.RawMessage(envBytes)

	return result, nil
}