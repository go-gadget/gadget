// Copyright 2016 Ryan Boehning. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package j

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	maxLineWidth = 80
)

// logger writes pretty logs to the $TMPDIR/q file. It takes care of opening and
// closing the file. It is safe for concurrent use.
type logger struct {
	mu        sync.Mutex // protects all the other fields
	start     time.Time  // time of first write in the current log group
	lastWrite time.Time  // last time buffer was flushed. determines when to print header
	lastFile  string     // last file to call q.Q(). determines when to print header
	lastFunc  string     // last function to call q.Q(). determines when to print header
}

// header returns a formatted header string, e.g. [14:00:36 main.go main.main:122]
// if the 2s timer has expired, or the calling function or filename has changed.
// If none of those things are true, it returns an empty string.
func (l *logger) header(funcName, file string, line int) string {
	if !l.shouldPrintHeader(funcName, file) {
		return ""
	}

	now := time.Now().UTC()
	l.start = now
	l.lastFunc = funcName
	l.lastFile = file

	return fmt.Sprintf("[%s %s:%d %s]", now.Format("15:04:05"), shortFile(file), line, funcName)
}

func (l *logger) shouldPrintHeader(funcName, file string) bool {
	if file != l.lastFile {
		return true
	}

	if funcName != l.lastFunc {
		return true
	}

	// If less than 2s has elapsed, this log line will be printed under the
	// previous header.
	const timeWindow = 2 * time.Second

	return time.Since(l.lastWrite) > timeWindow
}

// output writes to the log buffer. Each log message is prepended with a
// timestamp. Long lines are broken at 80 characters.
func (l *logger) output(args ...string) {
	timestamp := fmt.Sprintf("%.3fs", time.Since(l.start).Seconds())
	timestampWidth := len(timestamp) + 1 // +1 for padding space after timestamp

	// preWidth is the length of everything before the log message.
	fmt.Print(timestamp, " ")

	// Subsequent lines have to be indented by the width of the timestamp.
	indent := strings.Repeat(" ", timestampWidth)
	padding := "" // padding is the space between args.
	lineArgs := 0 // number of args printed on the current log line.
	lineWidth := timestampWidth
	for _, arg := range args {
		argWidth := argWidth(arg)
		lineWidth += argWidth + len(padding)

		// Some names in name=value strings contain newlines. Insert indentation
		// after each newline so they line up.
		arg = strings.Replace(arg, "\n", "\n"+indent, -1)

		// Break up long lines. If this is first arg printed on the line
		// (lineArgs == 0), it makes no sense to break up the line.
		if lineWidth > maxLineWidth && lineArgs != 0 {
			fmt.Print("\n", indent)
			lineArgs = 0
			lineWidth = timestampWidth + argWidth
			padding = ""
		}
		fmt.Print(padding, arg)
		lineArgs++
		padding = " "
	}

	fmt.Println()
}

// shortFile takes an absolute file path and returns just the <directory>/<file>,
// e.g. "foo/bar.go".
func shortFile(file string) string {
	dir := filepath.Base(filepath.Dir(file))
	file = filepath.Base(file)
	return filepath.Join(dir, file)
}
