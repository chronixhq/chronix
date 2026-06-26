package execution

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/oliveagle/jsonpath"
	"gorm.io/datatypes"
)

// substituteVars replaces ${var} or {{var}} placeholders with their values from the map.
func substituteVars(text string, vars map[string]any) string {
	if vars == nil {
		return text
	}
	var sb strings.Builder
	curr := text
	for {
		b1 := strings.Index(curr, "${")
		b2 := strings.Index(curr, "{{")
		if b1 == -1 && b2 == -1 {
			sb.WriteString(curr)
			break
		}

		var start int
		var isDouble bool
		if b1 != -1 && (b2 == -1 || b1 < b2) {
			start = b1
			isDouble = false
		} else {
			start = b2
			isDouble = true
		}

		sb.WriteString(curr[:start])

		var end int
		var key string
		suffix := ""
		if isDouble {
			endOffset := strings.Index(curr[start+2:], "}}")
			if endOffset == -1 {
				sb.WriteString(curr[start : start+2])
				curr = curr[start+2:]
				continue
			}
			end = start + 2 + endOffset
			key = strings.TrimSpace(curr[start+2 : end])
			suffix = curr[end+2:]
		} else {
			endOffset := strings.Index(curr[start+2:], "}")
			if endOffset == -1 {
				sb.WriteString(curr[start : start+2])
				curr = curr[start+2:]
				continue
			}
			end = start + 2 + endOffset
			key = strings.TrimSpace(curr[start+2 : end])
			suffix = curr[end+1:]
		}

		if val, ok := vars[key]; ok {
			fmt.Fprint(&sb, val)
		} else {
			if isDouble {
				sb.WriteString(curr[start : end+2])
			} else {
				sb.WriteString(curr[start : end+1])
			}
		}
		curr = suffix
	}
	return sb.String()
}

func captureDatabaseVariables(capture *datatypes.JSONMap, rowsCount int, resultLines []map[string]any, _ map[string]any) map[string]any {
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
		case "column":
			name, _ := def["name"].(string)
			rowSel, _ := def["row"].(string)
			if rowsCount == 0 || len(resultLines) == 0 {
				continue
			}
			row := resultLines[0]
			if rowSel == "last" {
				row = resultLines[len(resultLines)-1]
			}
			if val, ok := row[name]; ok {
				newVars[varName] = val
			}
		case "jsonpath":
			name, _ := def["name"].(string)
			path, _ := def["path"].(string)
			rowSel, _ := def["row"].(string)
			if rowsCount == 0 || len(resultLines) == 0 {
				continue
			}
			row := resultLines[0]
			if rowSel == "last" {
				row = resultLines[len(resultLines)-1]
			}
			if val, ok := row[name]; ok {
				var jsonData any
				var b []byte
				switch t := val.(type) {
				case []byte:
					b = t
				case string:
					b = []byte(t)
				default:
					b, _ = json.Marshal(t)
				}
				if err := json.Unmarshal(b, &jsonData); err == nil {
					if res, err := jsonpath.JsonPathLookup(jsonData, path); err == nil {
						newVars[varName] = res
					}
				}
			}
		}
	}
	return newVars
}

func captureShellVariables(capture *datatypes.JSONMap, stdout []byte, stderr []byte, _ map[string]any) map[string]any {
	if capture == nil {
		return nil
	}
	newVars := map[string]any{}
	combined := string(stdout) + string(stderr)
	for varName, defRaw := range *capture {
		def, ok := defRaw.(map[string]any)
		if !ok {
			continue
		}
		source, _ := def["source"].(string)
		switch source {
		case "jsonpath":
			path, _ := def["path"].(string)
			var jsonData any
			if err := json.Unmarshal(stdout, &jsonData); err == nil {
				if res, err := jsonpath.JsonPathLookup(jsonData, path); err == nil {
					newVars[varName] = res
				}
			}
		case "regex":
			pattern, _ := def["pattern"].(string)
			groupRaw := def["group"]
			re, err := regexp.Compile(pattern)
			if err == nil {
				matches := re.FindStringSubmatch(combined)
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
