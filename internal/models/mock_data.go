package models

import (
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// Fixed mock data
// ---------------------------------------------------------------------------

// Version is the fixed version string returned by /api/version.
const Version = "0.5.7"

// GenerateReply is the fixed reply text of /api/generate.
const GenerateReply = "That is a great question!"

// ChatReply is the fixed reply text of /api/chat.
const ChatReply = "Hello! How can I help?"

// EmbeddingDim is the dimension of the fixed embedding vector.
const EmbeddingDim = 768

// EmbeddingValue is the value of every dimension of the fixed embedding vector.
const EmbeddingValue = 0.1

// GenerateTokens is GenerateReply split into streamed tokens. Concatenating all
// tokens reproduces the original text (the separator belongs to the next token).
var GenerateTokens = []string{"That", " is", " a", " great", " question!"}

// ChatTokens is ChatReply split into streamed tokens. Concatenating all tokens
// reproduces the original text.
var ChatTokens = []string{"Hello", "!", " How", " can", " I", " help", "?"}

// Fixed model list: llama3.2:latest, nomic-embed-text:latest, gemma2:2b.
var fixedModels = []ModelInfo{
	{
		Name:       "llama3.2:latest",
		Model:      "llama3.2:latest",
		ModifiedAt: "2025-01-15T10:30:00.000000000Z",
		Size:       2019393189,
		Digest:     "a006de4b1e2a4d3a8b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f",
		Details: ModelDetails{
			Format:            "gguf",
			Family:            "llama",
			Families:          []string{"llama"},
			ParameterSize:     "3.2B",
			QuantizationLevel: "Q4_K_M",
		},
	},
	{
		Name:       "nomic-embed-text:latest",
		Model:      "nomic-embed-text:latest",
		ModifiedAt: "2025-01-10T08:20:00.000000000Z",
		Size:       274301056,
		Digest:     "b007de4b1e2a4d3a8b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f",
		Details: ModelDetails{
			Format:            "gguf",
			Family:            "nomic-bert",
			Families:          []string{"nomic-bert"},
			ParameterSize:     "137M",
			QuantizationLevel: "F16",
		},
	},
	{
		Name:       "gemma2:2b",
		Model:      "gemma2:2b",
		ModifiedAt: "2025-02-01T14:45:00.000000000Z",
		Size:       1622813184,
		Digest:     "c008de4b1e2a4d3a8b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f",
		Details: ModelDetails{
			Format:            "gguf",
			Family:            "gemma2",
			Families:          []string{"gemma2"},
			ParameterSize:     "2B",
			QuantizationLevel: "Q4_K_M",
		},
	},
}

// Per-model fixed model_info, used by /api/show.
var fixedModelInfos = map[string]map[string]interface{}{
	"llama3.2:latest": {
		"general.architecture":       "llama",
		"llama.context_length":       131072,
		"llama.embedding_length":     3072,
		"llama.block_count":          28,
		"llama.attention.head_count": 32,
	},
	"nomic-embed-text:latest": {
		"general.architecture":            "nomic-bert",
		"nomic-bert.context_length":       8192,
		"nomic-bert.embedding_length":     768,
		"nomic-bert.block_count":          12,
		"nomic-bert.attention.head_count": 12,
	},
	"gemma2:2b": {
		"general.architecture":        "gemma2",
		"gemma2.context_length":       8192,
		"gemma2.embedding_length":     2304,
		"gemma2.block_count":          26,
		"gemma2.attention.head_count": 8,
	},
}

// Per-model fixed template, used by /api/show.
var fixedTemplates = map[string]string{
	"llama3.2:latest":         "{{ .Prompt }}",
	"nomic-embed-text:latest": "{{ .Prompt }}",
	"gemma2:2b":               "{{ .Prompt }}",
}

// Per-model fixed parameters, used by /api/show.
var fixedParameters = map[string]string{
	"llama3.2:latest":         `stop "<|start_header_id|>"`,
	"nomic-embed-text:latest": `stop "<|endoftext|>"`,
	"gemma2:2b":               `stop "<end_of_turn>"`,
}

// fixedLicense is the fixed license text returned by /api/show.
const fixedLicense = "MIT License\n\nCopyright (c) 2025 ollama-mock\n\nPermission is hereby granted, free of charge, to any person obtaining a copy " +
	"of this software and associated documentation files (the \"Software\"), to deal in the Software without restriction."

// FixedModels returns a copy of the fixed model list so callers cannot mutate
// the internal state.
func FixedModels() []ModelInfo {
	out := make([]ModelInfo, len(fixedModels))
	copy(out, fixedModels)
	return out
}

// FindModel looks up a fixed model by name and returns nil when it is not found.
// The ":latest" suffix may be omitted, e.g. "llama3.2" equals "llama3.2:latest".
func FindModel(name string) *ModelInfo {
	if name == "" {
		return nil
	}
	for i := range fixedModels {
		if fixedModels[i].Name == name || fixedModels[i].Model == name {
			return &fixedModels[i]
		}
	}
	// Retry with the ":latest" suffix appended, so a bare "llama3.2" also matches.
	if !strings.Contains(name, ":") {
		withTag := name + ":latest"
		for i := range fixedModels {
			if fixedModels[i].Name == withTag {
				return &fixedModels[i]
			}
		}
	}
	return nil
}

// FixedEmbedding returns the fixed 768-dimension vector whose every dimension is 0.1.
func FixedEmbedding() []float32 {
	v := make([]float32, EmbeddingDim)
	for i := range v {
		v[i] = EmbeddingValue
	}
	return v
}

// ---------------------------------------------------------------------------
// Fixed statistics
// ---------------------------------------------------------------------------

// GenerateStats returns the fixed statistics of /api/generate.
func GenerateStats() (total, load int64, promptEval, eval int) {
	return 600000000, 100000000, 12, 6
}

// ChatStats returns the fixed statistics of /api/chat.
func ChatStats() (total int64, promptEval, eval int) {
	return 700000000, 8, 7
}

// EmbedStats returns the fixed statistics of /api/embed, where prompt_eval_count
// equals the number of inputs.
func EmbedStats(inputCount int) (total, load int64, promptEval int) {
	return 1000000, 500000, inputCount
}

// ---------------------------------------------------------------------------
// /api/show fixed content
// ---------------------------------------------------------------------------

// ShowInfo returns the fixed /api/show payload of the given model.
func ShowInfo(name string) (ShowResponse, bool) {
	m := FindModel(name)
	if m == nil {
		return ShowResponse{}, false
	}

	modelInfo, ok := fixedModelInfos[m.Name]
	if !ok {
		modelInfo = map[string]interface{}{
			"general.architecture": m.Details.Family,
		}
	}
	template, ok := fixedTemplates[m.Name]
	if !ok {
		template = "{{ .Prompt }}"
	}
	parameters, ok := fixedParameters[m.Name]
	if !ok {
		parameters = "stop \"<|start_header_id|>\""
	}

	return ShowResponse{
		License:    fixedLicense,
		Modelfile:  "# Modelfile generated by mock\nFROM " + m.Name + "\n",
		Parameters: parameters,
		Template:   template,
		Details:    m.Details,
		ModelInfo:  modelInfo,
	}, true
}

// ---------------------------------------------------------------------------
// /api/ps fixed content
// ---------------------------------------------------------------------------

// ProcessModels returns the fixed list of "running" models.
func ProcessModels() []ProcessModelInfo {
	all := FixedModels()
	out := make([]ProcessModelInfo, 0, len(all))
	for _, m := range all {
		out = append(out, ProcessModelInfo{
			Name:       m.Name,
			Model:      m.Model,
			ModifiedAt: m.ModifiedAt,
			Size:       m.Size,
			Digest:     m.Digest,
			Details:    m.Details,
			ExpiresAt:  time.Now().UTC().Add(5 * time.Minute).Format(time.RFC3339Nano),
			SizeVram:   m.Size,
		})
	}
	return out
}
