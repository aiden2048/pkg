package errorMsg

import "testing"

func TestError2str(t *testing.T) {
	t.Log(NewError(1, "hello").Line("A").Error())
	t.Log(NewError(1, "hello").Line().Error())
	t.Log(NewError(1, "hello").Return("A").Error())
	t.Log(NewError(1, "hello").Return("").Error())
}
