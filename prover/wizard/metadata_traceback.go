package wizard

import (
	"fmt"
	"path"
	"runtime"
)

var (
	// currSourceDir is used to trim the frames that we put in metadata.
	currSourceDir = currSourceDirectory()
)

// getTraceBackFrames returns the runtime frame of all the callers except for
// the current function and at most 20 of them.
func getTraceBackFrames(skip, numFrame int) []runtime.Frame {

	var (
		frameList = make([]runtime.Frame, numFrame)
		// a pc can expand into more than one frame. The use of "numFrame" here
		// is an upper-bound of what is actually needed. The pcs don't account
		// for inlined functions.
		pcs = make([]uintptr, numFrame)
		// The first frame always come from the runtime package, so we may just
		// skip it.
		n = runtime.Callers(skip+1, pcs)
	)

	if n == 0 {
		// No pcs available. Stop now.
		// This can happen if the first argument to runtime.Callers is large.
		return nil
	}

	pcs = pcs[:n]
	frames := runtime.CallersFrames(pcs)

	for i := 0; i < numFrame; i++ {
		var more bool
		frameList[i], more = frames.Next()
		if !more {
			return frameList[:i+1]
		}
	}

	return frameList
}

// currSourceDirectory returns the source directory of the caller
func currSourceDirectory() string {
	frame := getTraceBackFrames(1, 1)[0]
	return path.Dir(frame.File)
}

// trimFrames removes the initial frames that come from the current directory
// and then keep all the the frames up to the point where we get back to the
// current directory or the end of the list of frames.
func trimFramesFromCurrDir(frames []runtime.Frame) []runtime.Frame {

	var (
		indexFirstFoundNotCurrDir int = -1
	)

	for currIndex := range frames {

		var (
			srcDir        = path.Dir(frames[currIndex].File)
			isFromCurrDir = srcDir == currSourceDir
		)

		if !isFromCurrDir && indexFirstFoundNotCurrDir < 0 {
			indexFirstFoundNotCurrDir = currIndex
		}

		if isFromCurrDir && indexFirstFoundNotCurrDir >= 0 {
			return frames[indexFirstFoundNotCurrDir:currIndex]
		}
	}

	if indexFirstFoundNotCurrDir < 0 {
		return nil
	}

	return frames[indexFirstFoundNotCurrDir:]
}

// formatFrames format a list of frames as file:line.(func) and returns a list
// of string.
func formatFrames(frames []runtime.Frame) []string {
	res := make([]string, len(frames))
	for i := range frames {
		res[i] = fmt.Sprintf("%v:%v.(%v)", frames[i].File, frames[i].Line, frames[i].Function)
	}
	return res
}
