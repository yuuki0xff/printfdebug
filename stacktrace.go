package main

import (
	"runtime"
	"fmt"
	"runtime/debug"
	"os"
)

func printDebug() {
	maxStackSize:=1024
	skips := 2
	pc := make([]uintptr, maxStackSize)
	pclen := runtime.Callers(skips, pc)
	pc = pc[:pclen]

	s := fmt.Sprintf("StackTrace\n")
	frames := runtime.CallersFrames(pc);
	for {
		frame, more := frames.Next()
		if more == false {
			break
		}

		s += fmt.Sprintf("  %s:%d (%s) PC=%d, func=%s entry=%d\n",
			frame.File, frame.Line, frame.Function,
			frame.PC, frame.Func, frame.Entry)
	}
	print(s)
	os.Stdout.Write(debug.Stack())
}

func b(args string, i int) {
	println("start b")
	printDebug()
	println("end b")
}

func a(args string) {
	b(args, 0)
}

func main() {
	a("test")
}

