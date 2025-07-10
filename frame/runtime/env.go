package runtime

import (
	"os"
	"strings"
)

var isForceDebug, isDebug, isDev bool

func init() {
	for i := 0; i < len(os.Args); i++ {
		//	fmt.Printf("flag.Arg(%d) = %s\n", i, flag.Arg(i))
		if strings.ToLower(os.Args[i]) == "-d" || strings.ToLower(os.Args[i]) == "dbg" {
			isDebug = true
			break
		}
	}

	for i := 0; i < len(os.Args); i++ {
		//	fmt.Printf("flag.Arg(%d) = %s\n", i, flag.Arg(i))
		if strings.ToLower(os.Args[i]) == "dev" {
			isDev = true
			break
		}
	}
}

func SetForceDebug(force bool) {
	isForceDebug = force
}

func IsDebug() bool {
	if isForceDebug {
		return true
	}
	return isDebug
}

func IsDev() bool {
	return isDev
}
