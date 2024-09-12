package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type StackTrace struct {
	GoroutineNumber int
	GoroutineState  string
	StackFrames     []StackFrame
}

func (st StackTrace) String() string {
	var res strings.Builder
	fmt.Fprintf(&res, "goroutine %d [%s]:\n", st.GoroutineNumber, st.GoroutineState)
	for _, sf := range st.StackFrames {
		fmt.Fprintf(&res, "%s\n", sf)
	}
	return res.String()
}

func (st StackTrace) Equal(st2 StackTrace) bool {
	if len(st.StackFrames) != len(st2.StackFrames) {
		return false
	}
	for i := range st.StackFrames {
		if st.StackFrames[i] != st2.StackFrames[i] {
			return false
		}
	}
	return true
}

type StackFrame struct {
	FunctionName string
	FileName     string
	FileLine     int
}

func (sf StackFrame) String() string {
	var res strings.Builder
	fmt.Fprintf(&res, "%s\n\t%s:%d", sf.FunctionName, sf.FileName, sf.FileLine)
	return res.String()
}

type parserState string

const (
	stateGoroutine    parserState = "goroutine"
	stateFunctionName parserState = "functionName"
	stateFileLine     parserState = "fileLine"
	stateSpacer       parserState = "spacer"
)

func Parse(r io.Reader) ([]StackTrace, error) {
	s := bufio.NewScanner(r)

	var res []StackTrace

	for {
		st, err := parseGoroutine(s)
		if err != nil {
			return nil, fmt.Errorf("failed to parse goroutine: %w", err)
		}
		if st == nil {
			break
		}
		res = append(res, *st)
	}

	return res, nil
}

func parseGoroutine(s *bufio.Scanner) (*StackTrace, error) {
	var st *StackTrace
	var i int
	var prevLine string

	state := stateGoroutine
	for s.Scan() {
		origLine := s.Text()
		line := strings.TrimSpace(origLine)
		if line == "" {
			state = stateSpacer
		}

		if strings.HasPrefix(line, "exit status") {
			continue
		}

		switch state {
		case stateGoroutine:
			line = strings.TrimPrefix(line, "goroutine ")
			number, rest, found := strings.Cut(line, " ")
			if !found {
				return nil, fmt.Errorf("unexpected line: %s, previous line: %s, state=%s", origLine, prevLine, state)
			}
			groutineNumber, err := strconv.Atoi(number)
			if err != nil {
				return nil, fmt.Errorf("unexpected line: %s, previous line: %s, state=%s, %w", origLine, prevLine, state, err)
			}

			goroutineState := rest[1 : len(rest)-2]
			todo := strings.Split(goroutineState, ", ")
			goroutineState = todo[0]

			st = &StackTrace{
				GoroutineNumber: groutineNumber,
				GoroutineState:  goroutineState,
				StackFrames:     make([]StackFrame, 0),
			}

			state = stateFunctionName
		case stateFunctionName:
			functionName := strings.TrimSuffix(line, "(...)")
			functionName = strings.TrimSuffix(functionName, "()")
			functionName, _, _ = strings.Cut(functionName, "(0x")
			functionName, _, _ = strings.Cut(functionName, "({0x")
			functionName, _, _ = strings.Cut(functionName, "({{0x")

			st.StackFrames = append(st.StackFrames, StackFrame{FunctionName: functionName})

			state = stateFileLine
		case stateFileLine:
			filename, rest, found := strings.Cut(line, ":")
			if !found {
				return nil, fmt.Errorf("unexpected line: %s, previous line: %s, state=%s", origLine, prevLine, state)
			}
			st.StackFrames[i].FileName = filename

			fields := strings.Fields(rest)
			fileLine, err := strconv.Atoi(fields[0])
			if err != nil {
				return nil, fmt.Errorf("unexpected line: %s, previous line: %s, state=%s, %w", origLine, prevLine, state, err)
			}
			st.StackFrames[i].FileLine = fileLine

			i++
			state = stateFunctionName
		case stateSpacer:
			return st, nil
		}
		prevLine = origLine
	}

	return st, nil
}
