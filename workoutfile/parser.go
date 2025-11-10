package workoutfile

import (
	"fmt"
	"path/filepath"
	"strings"
	"zone-finder/fit"
	"zone-finder/tcx"
)

func ParseFile(path string) (WorkoutFile, error) {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".tcx":
		return tcx.ParseTCX(path)
	case ".fit":
		return fit.ParseFIT(path)
	default:
		return nil, fmt.Errorf("unsupported file format %s", ext)
	}
}
