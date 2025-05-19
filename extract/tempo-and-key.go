package extract

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

func ExtractTempoAndKey(filename string) (string, int, error) {
	cmd := exec.Command("python3", "extract/tempo-and-key.py", filename)
	cmd.Stderr = nil
	output, err := cmd.Output()
	if err != nil {
		return "", 0, fmt.Errorf("error running Python script: %w - output: %s", err, string(output))
	}
	var result struct {
		Key string `json:"key"`
		BPM int    `json:"bpm"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return "", 0, fmt.Errorf("error parsing JSON: %w", err)
	}

	return result.Key, result.BPM, nil
}
