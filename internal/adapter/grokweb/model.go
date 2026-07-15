package grokweb

import "strings"

type modelSpec struct {
	ModelName string
	ModelMode string
}

var supportedModels = []string{
	"grok-3",
	"grok-3-mini",
	"grok-3-thinking",
	"grok-4",
	"grok-4-mini",
	"grok-4-thinking",
	"grok-4-heavy",
	"grok-4.1-mini",
	"grok-4.1-fast",
	"grok-4.1-expert",
	"grok-4.1-thinking",
}

var modelMap = map[string]modelSpec{
	"grok-3":            {ModelName: "grok-3", ModelMode: "MODEL_MODE_GROK_3"},
	"grok-3-mini":       {ModelName: "grok-3", ModelMode: "MODEL_MODE_GROK_3_MINI_THINKING"},
	"grok-3-thinking":   {ModelName: "grok-3", ModelMode: "MODEL_MODE_GROK_3_THINKING"},
	"grok-4":            {ModelName: "grok-4", ModelMode: "MODEL_MODE_GROK_4"},
	"grok-4-mini":       {ModelName: "grok-4-mini", ModelMode: "MODEL_MODE_GROK_4_MINI_THINKING"},
	"grok-4-thinking":   {ModelName: "grok-4", ModelMode: "MODEL_MODE_GROK_4_THINKING"},
	"grok-4-heavy":      {ModelName: "grok-4", ModelMode: "MODEL_MODE_HEAVY"},
	"grok-4.1-mini":     {ModelName: "grok-4-1-thinking-1129", ModelMode: "MODEL_MODE_GROK_4_1_MINI_THINKING"},
	"grok-4.1-fast":     {ModelName: "grok-4-1-thinking-1129", ModelMode: "MODEL_MODE_FAST"},
	"grok-4.1-expert":   {ModelName: "grok-4-1-thinking-1129", ModelMode: "MODEL_MODE_EXPERT"},
	"grok-4.1-thinking": {ModelName: "grok-4-1-thinking-1129", ModelMode: "MODEL_MODE_GROK_4_1_THINKING"},
}

func lookupModel(clientModel string) (modelSpec, bool) {
	clientModel = strings.TrimPrefix(clientModel, "grok-web/")
	spec, ok := modelMap[clientModel]
	return spec, ok
}
