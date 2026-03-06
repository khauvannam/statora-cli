package extcmd

import "fmt"

func errNoActiveVersion() error {
	return fmt.Errorf("no active PHP version — run `statora switch` first")
}
