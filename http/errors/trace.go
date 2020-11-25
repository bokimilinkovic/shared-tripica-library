package errors

import (
	"runtime"
	"strings"
)

const (
	// Second level call means that we need the name of a function which indirectly calls withTrace.
	secondLevelTraceFunctionCall = 2

	// Third level call means that we are in a helper function, and we need the function calling it, instead.
	thirdLevelTraceFunctionCall = 3
)

// withTrace returns the name of the function.
// For example, if stackFramesToSkip is equal to 2, then:
// if funcA() calls funcB(), and funcB() calls withTrace(1) => "funcB" is returned
// if funcA() calls funcB(), and funcB() calls withTrace(2) => "funcA" is returned.
func wthTrace(stackFramesToSkip int) string {
	pc, _, _, ok := runtime.Caller(stackFramesToSkip)
	if !ok {
		return ""
	}

	fn := runtime.FuncForPC(pc)
	return fn.Name()
}

// isHelperFunction determines if the function is equal to one of the possible helper functions in the project.
// For now, it's only service/collectai's handleHTTPError function, but others can be added if needed.
// This is a not so pretty function, and tests should be able to catch if something changes which breaks the check.
func isHelperFunction(fn string) bool {
	return strings.HasSuffix(fn, "handleHTTPError")
}
