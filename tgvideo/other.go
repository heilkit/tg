package tgvideo

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

// ExtractMetadata gets you metadata at format/tags/comment.
// It's sort of convenient, taking in mind you can't have custom metadata keys with .mp4
func ExtractMetadata[T any](filename string, opts ...*Opt) (T, error) {
	var ret T

	cmdRet, err := ExtractMetadataAll[metadataCommandReturnCut](filename, opts...)
	if err != nil {
		return ret, err
	}

	metadata, contains := cmdRet.Format.Tags["comment"]
	if !contains {
		return ret, nil
	}

	if err := json.Unmarshal([]byte(metadata), &ret); err != nil {
		return ret, fmt.Errorf("while parsing metadata format/tags/comment: %v", err)
	}
	return ret, nil
}

// ExtractMetadataAll gets you all metadata you would need.
// If you need EVERY single key-value, use ExtractMetadataAll[map[string]any](...).
func ExtractMetadataAll[T any](filename string, opts ...*Opt) (*T, error) {
	options := parseOpts(opts...)

	output, err := exec.Command(options.Ffprobe, filename, "-print_format", "json", "-show_format").
		Output()
	if err != nil {
		return nil, fmt.Errorf("while executing %s: %v", options.Ffprobe, err)
	}

	var ret T
	if err := json.Unmarshal(output, &ret); err != nil {
		return nil, fmt.Errorf("while parsing video metadata: %v", err)
	}

	return &ret, nil
}

type metadataCommandReturnCut struct {
	Format struct {
		Tags map[string]string `json:"tags"`
	} `json:"format"`
}
