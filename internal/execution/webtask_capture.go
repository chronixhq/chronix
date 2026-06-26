package execution

import (
	"encoding/json"
	"regexp"

	"chronix/internal/webtaskrun"

	"github.com/oliveagle/jsonpath"
	"gorm.io/datatypes"
)

func captureWebTaskVariables(capture *datatypes.JSONMap, result *webtaskrun.WebTaskResult, _ map[string]any) map[string]any {
	if capture == nil {
		return nil
	}
	newVars := map[string]any{}
	for varName, defRaw := range *capture {
		def, ok := defRaw.(map[string]any)
		if !ok {
			continue
		}
		source, _ := def["source"].(string)
		switch source {
		case "jsonpath":
			path, _ := def["path"].(string)
			var jsonData interface{}
			if err := json.Unmarshal(result.ResponseBody, &jsonData); err == nil {
				if res, err := jsonpath.JsonPathLookup(jsonData, path); err == nil {
					newVars[varName] = res
				}
			}
		case "header":
			name, _ := def["name"].(string)
			if val := result.ResponseHeaders.Get(name); val != "" {
				newVars[varName] = val
			}
		case "regex":
			pattern, _ := def["pattern"].(string)
			groupRaw := def["group"]
			re, err := regexp.Compile(pattern)
			if err == nil {
				matches := re.FindStringSubmatch(string(result.ResponseBody))
				if len(matches) > 0 {
					group := 0
					if groupRaw != nil {
						group = intFromAny(groupRaw)
					} else if len(matches) > 1 {
						group = 1
					}
					if len(matches) > group {
						newVars[varName] = matches[group]
					}
				}
			}
		}
	}
	return newVars
}
