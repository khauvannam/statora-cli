package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const zshHook = `autoload -U add-zsh-hook

_statora_use() {
  if command -v statora >/dev/null 2>&1; then
    eval "$(statora use --shell zsh 2>/dev/null)"
  fi
}

add-zsh-hook chpwd _statora_use
_statora_use
`

const bashHook = `_statora_use() {
  if command -v statora >/dev/null 2>&1; then
    eval "$(statora use --shell bash 2>/dev/null)"
  fi
}

if [[ "${PROMPT_COMMAND}" != *"_statora_use"* ]]; then
  PROMPT_COMMAND="_statora_use${PROMPT_COMMAND:+;$PROMPT_COMMAND}"
fi
_statora_use
`

const fishHook = `function __statora_use --on-variable PWD
  if command -q statora
    statora use --shell fish 2>/dev/null | source
  end
end
__statora_use
`

var hooks = map[string]string{
	"zsh":  zshHook,
	"bash": bashHook,
	"fish": fishHook,
}

// DetectShell resolves the shell name from the SHELL env var and optional flag override.
// Returns an error if the shell is unsupported.
func DetectShell(shellEnv, flagVal string) (string, error) {
	s := flagVal
	if s == "" {
		s = filepath.Base(shellEnv)
	}
	if _, ok := hooks[s]; ok {
		return s, nil
	}
	return "", fmt.Errorf("unsupported shell %q — use --shell zsh|bash|fish", s)
}

// PrintEnv writes the shell hook snippet for the given shell to w.
func PrintEnv(w io.Writer, shell string) error {
	hook, ok := hooks[shell]
	if !ok {
		return fmt.Errorf("unsupported shell %q", shell)
	}
	_, err := fmt.Fprint(w, hook)
	return err
}

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Print shell hook for auto-switching PHP/Composer on directory change",
	Long: `Print shell integration code for auto-switching.

Add to your shell rc file:

  zsh/bash:  eval "$(statora env)"
  fish:      statora env | source`,
}

func init() {
	var shellFlag string
	envCmd.Flags().StringVar(&shellFlag, "shell", "", "Shell to target (zsh, bash, fish)")
	envCmd.RunE = func(_ *cobra.Command, _ []string) error {
		shell, err := DetectShell(os.Getenv("SHELL"), shellFlag)
		if err != nil {
			return err
		}
		return PrintEnv(os.Stdout, shell)
	}
	Root.AddCommand(envCmd)
}
