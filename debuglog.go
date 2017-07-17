package printfdebug

import (
	"encoding/json"
	"os"
	"runtime"
)

const (
	skips = 3
)

var (
	MaxStackSize = 1024
	OutputFile   = os.Stdout
)

type LogMessage struct {
	Tag    string          `json:"tag"`
	Frames []runtime.Frame `json:"frames"`
}

func printDebugMsg(tag string) {
	logmsg := LogMessage{}
	logmsg.Tag = tag
	logmsg.Frames = make([]runtime.Frame, 0, MaxStackSize)

	pc := make([]uintptr, MaxStackSize)
	pclen := runtime.Callers(skips, pc)
	pc = pc[:pclen]

	frames := runtime.CallersFrames(pc)
	for {
		frame, more := frames.Next()
		if more == false {
			break
		}

		logmsg.Frames = append(logmsg.Frames, frame)
	}
	js, err := json.Marshal(logmsg)
	if err != nil {
		panic(err)
	}
	OutputFile.Write(js)
	OutputFile.Write([]byte("\n"))
}

func FuncStart() {
	printDebugMsg("funcStart")
}

func FuncEnd() {
	printDebugMsg("funcEnd")
}
