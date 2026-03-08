package use

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// StripStatoraDirs removes ~/.statora/... entries from a colon-separated PATH string.
func StripStatoraDirs(path, homeDir string) string {
	statoraPrefix := homeDir + "/.statora/"
	var kept []string
	for p := range strings.SplitSeq(path, ":") {
		if p != "" && !strings.HasPrefix(p, statoraPrefix) {
			kept = append(kept, p)
		}
	}
	return strings.Join(kept, ":")
}

// BuildPATHExport returns the shell-specific line that sets PATH to prepend
// phpBinDir and composerBinDir in front of strippedPATH.
func BuildPATHExport(shell, phpBinDir, composerBinDir, strippedPATH string) string {
	switch shell {
	case "fish":
		dirs := []string{phpBinDir, composerBinDir}
		if strippedPATH != "" {
			dirs = append(dirs, strings.Split(strippedPATH, ":")...)
		}
		return fmt.Sprintf("set -gx PATH %s\n", strings.Join(dirs, " "))
	default:
		newPath := phpBinDir + ":" + composerBinDir
		if strippedPATH != "" {
			newPath += ":" + strippedPATH
		}
		return fmt.Sprintf("export PATH=%q\n", newPath)
	}
}

// PrintUse writes the PATH export line for the given shell and resolved dirs to w.
func PrintUse(w io.Writer, shell, phpBinDir, composerBinDir, currentPATH, homeDir string) error {
	stripped := StripStatoraDirs(currentPATH, homeDir)
	_, err := fmt.Fprint(w, BuildPATHExport(shell, phpBinDir, composerBinDir, stripped))
	return err
}

// IsTerminal reports whether fd is connected to a terminal (for interactive prompts).
func IsTerminal(fd *os.File) bool {
	info, err := fd.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}
