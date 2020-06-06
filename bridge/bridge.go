package bridge

type Functions interface {
	Print(str string)
	Log(str string, level string)
}
var Func Functions
