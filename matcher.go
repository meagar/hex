package hex

import (
	"fmt"
	"net/http"
)

type matcher interface {
	matches(req *http.Request) bool
	String() string
}

func matcherArgsToString(args []interface{}) string {
	if len(args) == 0 {
		return "<no args>"
	}

	if len(args) == 1 {
		return fmt.Sprintf("%s", args[0])
	}

	if len(args) == 2 {
		return fmt.Sprintf("%s=%q", args[0], args[1])
	}

	panic("Too many arguments")
}
