package bridge

import "fmt"

type Functions interface {
	Print(str string)
	Log(str string, level string)
}
var Func Functions

func Printf(format string, a ...interface{}) {
	Func.Print(fmt.Sprintf(format, a...))
}