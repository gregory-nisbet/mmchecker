package core

import (
	"fmt"
	"os"
	"strings"
)

var Verbosity int

func Vprint(vlevel int, args ...string) {
	if Verbosity >= vlevel {
		fmt.Fprintf(os.Stderr, "%s\n", strings.Join(args, " "))
	}
}
