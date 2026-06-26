package updater

import (
	"strconv"
	"strings"
)

func IsNewerVersion(current, latest string) bool {
	if current == "" || latest == "" {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(current), "dev") {
		return false
	}
	return compareVersions(current, latest) < 0
}

func compareVersions(a, b string) int {
	aParts, aOK := parseVersionParts(a)
	bParts, bOK := parseVersionParts(b)
	if !aOK || !bOK {
		return strings.Compare(strings.TrimSpace(a), strings.TrimSpace(b))
	}

	maxLen := len(aParts)
	if len(bParts) > maxLen {
		maxLen = len(bParts)
	}
	for i := 0; i < maxLen; i++ {
		var av, bv int
		if i < len(aParts) {
			av = aParts[i]
		}
		if i < len(bParts) {
			bv = bParts[i]
		}
		if av < bv {
			return -1
		}
		if av > bv {
			return 1
		}
	}
	return 0
}

func parseVersionParts(v string) ([]int, bool) {
	v = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(v, "v"), "V"))
	if v == "" {
		return nil, false
	}

	core := v
	if idx := strings.IndexAny(core, "-+"); idx >= 0 {
		core = core[:idx]
	}

	rawParts := strings.Split(core, ".")
	parts := make([]int, 0, len(rawParts))
	for _, part := range rawParts {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, false
		}
		digits := make([]rune, 0, len(part))
		for _, r := range part {
			if r < '0' || r > '9' {
				break
			}
			digits = append(digits, r)
		}
		if len(digits) == 0 {
			return nil, false
		}
		n, err := strconv.Atoi(string(digits))
		if err != nil {
			return nil, false
		}
		parts = append(parts, n)
	}
	return parts, true
}
