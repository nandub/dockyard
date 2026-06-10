package cli

import (
	"fmt"
	"os"
	"strings"
)

func osReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func printSimpleDiff(oldText string, newText string) {
	oldLines := strings.Split(oldText, "\n")
	newLines := strings.Split(newText, "\n")
	max := len(oldLines)
	if len(newLines) > max {
		max = len(newLines)
	}
	for i := 0; i < max; i++ {
		var oldLine, newLine string
		if i < len(oldLines) {
			oldLine = oldLines[i]
		}
		if i < len(newLines) {
			newLine = newLines[i]
		}
		if oldLine == newLine {
			continue
		}
		if oldLine != "" {
			fmt.Printf("-%s\n", oldLine)
		}
		if newLine != "" {
			fmt.Printf("+%s\n", newLine)
		}
	}
}
