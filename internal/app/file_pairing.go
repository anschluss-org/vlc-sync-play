package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FilePair represents a pair of conductor and audience files
type FilePair struct {
	ConductorFile string
	AudienceFile  string
}

// FindFilePair attempts to find a matching file pair based on naming convention
// Naming convention: [basename]-conductor.[ext] and [basename]-audience.[ext]
// If a complete pair is found, returns it. Otherwise returns the original paths.
func FindFilePair(inputPaths []string) ([]string, error) {
	if len(inputPaths) == 0 {
		return inputPaths, nil
	}

	// If we already have 2 or more files, use them as-is
	if len(inputPaths) >= 2 {
		return inputPaths, nil
	}

	// We have exactly one file - try to find its pair
	inputFile := inputPaths[0]

	// Parse the input file
	dir := filepath.Dir(inputFile)
	filename := filepath.Base(inputFile)
	ext := filepath.Ext(filename)
	nameWithoutExt := strings.TrimSuffix(filename, ext)

	// Check if it matches our naming convention
	var basename string
	var variant string

	if strings.HasSuffix(nameWithoutExt, "-conductor") {
		basename = strings.TrimSuffix(nameWithoutExt, "-conductor")
		variant = "conductor"
	} else if strings.HasSuffix(nameWithoutExt, "-audience") {
		basename = strings.TrimSuffix(nameWithoutExt, "-audience")
		variant = "audience"
	} else {
		// Doesn't match convention - try both variants
		basename = nameWithoutExt
		variant = ""
	}

	var conductorFile, audienceFile string

	if variant == "conductor" {
		conductorFile = inputFile
		audienceFile = filepath.Join(dir, basename+"-audience"+ext)
	} else if variant == "audience" {
		audienceFile = inputFile
		conductorFile = filepath.Join(dir, basename+"-conductor"+ext)
	} else {
		// No variant in name - look for both
		conductorFile = filepath.Join(dir, basename+"-conductor"+ext)
		audienceFile = filepath.Join(dir, basename+"-audience"+ext)
	}

	// Check if both files exist
	conductorExists := fileExists(conductorFile)
	audienceExists := fileExists(audienceFile)

	if conductorExists && audienceExists {
		// Found a complete pair!
		return []string{conductorFile, audienceFile}, nil
	}

	// If we started with a variant file but couldn't find its pair, show a helpful message
	if variant != "" {
		var missingFile string
		if !conductorExists {
			missingFile = conductorFile
		} else {
			missingFile = audienceFile
		}
		return inputPaths, fmt.Errorf("found %s but could not find matching file: %s", variant, missingFile)
	}

	// No pair found, use the original file (will fall back to same file in both instances)
	return inputPaths, nil
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
