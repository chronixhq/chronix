package main

import (
	"fmt"
	"strconv"
	"strings"
)

func bumpVersion(currentVersion, bumpType string) (string, error) {
	parts := strings.Split(currentVersion, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid version format: %s", currentVersion)
	}
	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	maint, _ := strconv.Atoi(parts[2])

	switch bumpType {
	case "maint":
		maint++
	case "minor":
		minor++
		maint = 0
	case "major":
		major++
		minor = 0
		maint = 0
	default:
		return "", fmt.Errorf("invalid bump type: %s", bumpType)
	}
	return fmt.Sprintf("%d.%d.%d", major, minor, maint), nil
}
