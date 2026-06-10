package config

import (
	"fmt"
	"regexp"
)

var forbiddenMeta = regexp.MustCompile("[;`$\n\r<>|]")

func validateCommand(s string) error {
	if forbiddenMeta.MatchString(s) {
		return fmt.Errorf("command contains forbidden metacharacters: %q", s)
	}
	return nil
}
