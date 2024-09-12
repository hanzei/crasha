package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStackTraceEqual(t *testing.T) {
	for name, tc := range map[string]struct {
		st1           StackTrace
		st2           StackTrace
		shouldBeEqual bool
	}{
		"both empty": {
			st1:           StackTrace{},
			st2:           StackTrace{},
			shouldBeEqual: true,
		},
		"one stack frame": {
			st1: StackTrace{
				GoroutineNumber: 1,
				GoroutineState:  "running",
				StackFrames: []StackFrame{
					{
						FunctionName: "reflect.mapiternext",
						FileName:     "/usr/local/go/src/runtime/map.go",
						FileLine:     1532,
					},
				},
			},
			st2: StackTrace{
				GoroutineNumber: 2,
				GoroutineState:  "runnable",
				StackFrames: []StackFrame{
					{
						FunctionName: "reflect.mapiternext",
						FileName:     "/usr/local/go/src/runtime/map.go",
						FileLine:     1532,
					},
				},
			},
			shouldBeEqual: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.shouldBeEqual, tc.st1.Equal(tc.st2))
		})
	}
}

func TestParse(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		r := strings.NewReader("")
		st, err := Parse(r)
		require.NoError(t, err)
		require.Len(t, st, 0)
	})

	t.Run("single goroutine", func(t *testing.T) {
		input := `goroutine 1 [running]:
reflect.mapiternext(0x47d3a9?)
	/usr/local/go/src/runtime/map.go:1532 +0x13
reflect.(*MapIter).Next(0xc000143878?)
	/usr/local/go/src/reflect/value.go:1989 +0x74
internal/fmtsort.Sort({0x4a6b80?, 0xc00012c120?, 0xc000143ad0?})
	/usr/local/go/src/internal/fmtsort/sort.go:59 +0x1a7
fmt.(*pp).printValue(0xc0001325b0, {0x4a6b80?, 0xc00012c120?, 0xc000143b68?}, 0x76, 0x0)
	/usr/local/go/src/fmt/print.go:816 +0x988
fmt.(*pp).printArg(0xc0001325b0, {0x4a6b80, 0xc00012c120}, 0x76)
	/usr/local/go/src/fmt/print.go:759 +0x4bb
fmt.(*pp).doPrintf(0xc0001325b0, {0x4ba9ac, 0x8}, {0xc000143ef8, 0x1, 0x1})
	/usr/local/go/src/fmt/print.go:1074 +0x37e
fmt.Appendf({0xc00014a020, 0x14, 0x20}, {0x4ba9ac, 0x8}, {0xc000143ef8, 0x1, 0x1})
	/usr/local/go/src/fmt/print.go:249 +0x7a
main.main.Printf.func3({0xc00014a020?, 0xc000134030?, 0x7a6bda0d4b88?})
	/usr/local/go/src/log/log.go:398 +0x2c
log.(*Logger).output(0xc00012c0f0, 0x0, 0x2, 0xc000143f08)
	/usr/local/go/src/log/log.go:238 +0x352
log.Printf(...)
	/usr/local/go/src/log/log.go:397
main.main()
	/home/bschumacher/src/tmp/maps/main.go:29 +0x165`

		r := strings.NewReader(input)
		st, err := Parse(r)
		require.NoError(t, err)
		require.Len(t, st, 1)
		require.Equal(t, 1, st[0].GoroutineNumber)
		require.Equal(t, "running", st[0].GoroutineState)
		require.Len(t, st[0].StackFrames, 11)

		assert.Equal(t, "reflect.mapiternext", st[0].StackFrames[0].FunctionName)
		assert.Equal(t, "/usr/local/go/src/runtime/map.go", st[0].StackFrames[0].FileName)
		assert.Equal(t, 1532, st[0].StackFrames[0].FileLine)
	})

	t.Run("multiple goroutines", func(t *testing.T) {
		input := `goroutine 1 [running]:
reflect.mapiternext(0x47d3a9?)
	/usr/local/go/src/runtime/map.go:1532 +0x13
reflect.(*MapIter).Next(0xc000143878?)
	/usr/local/go/src/reflect/value.go:1989 +0x74
internal/fmtsort.Sort({0x4a6b80?, 0xc00012c120?, 0xc000143ad0?})
	/usr/local/go/src/internal/fmtsort/sort.go:59 +0x1a7
fmt.(*pp).printValue(0xc0001325b0, {0x4a6b80?, 0xc00012c120?, 0xc000143b68?}, 0x76, 0x0)
	/usr/local/go/src/fmt/print.go:816 +0x988
fmt.(*pp).printArg(0xc0001325b0, {0x4a6b80, 0xc00012c120}, 0x76)
	/usr/local/go/src/fmt/print.go:759 +0x4bb
fmt.(*pp).doPrintf(0xc0001325b0, {0x4ba9ac, 0x8}, {0xc000143ef8, 0x1, 0x1})
	/usr/local/go/src/fmt/print.go:1074 +0x37e
fmt.Appendf({0xc00014a020, 0x14, 0x20}, {0x4ba9ac, 0x8}, {0xc000143ef8, 0x1, 0x1})
	/usr/local/go/src/fmt/print.go:249 +0x7a
main.main.Printf.func3({0xc00014a020?, 0xc000134030?, 0x7a6bda0d4b88?})
	/usr/local/go/src/log/log.go:398 +0x2c
log.(*Logger).output(0xc00012c0f0, 0x0, 0x2, 0xc000143f08)
	/usr/local/go/src/log/log.go:238 +0x352
log.Printf(...)
	/usr/local/go/src/log/log.go:397
main.main()
	/home/bschumacher/src/tmp/maps/main.go:29 +0x165

goroutine 18 [runnable]:
main.main.func1()
	/home/bschumacher/src/tmp/maps/main.go:16 +0x71
created by main.main in goroutine 1
	/home/bschumacher/src/tmp/maps/main.go:14 +0x87

goroutine 19 [running]:
main.main.func2()
	/home/bschumacher/src/tmp/maps/main.go:23 +0x65
created by main.main in goroutine 1
	/home/bschumacher/src/tmp/maps/main.go:22 +0xe7

goroutine 1450 [select, 2 minutes]:
runtime.gopark(0xc1d07d3958?, 0x2?, 0xc5?, 0x63?, 0xc1d07d38e4?)
	runtime/proc.go:381 +0xd6 fp=0xc1d07d3758 sp=0xc1d07d3738 pc=0x43d896
runtime.selectgo(0xc1d07d3958, 0xc1d07d38e0, 0xc1d07d38e8?, 0x0, 0xc1d07d38e8?, 0x1)
	runtime/select.go:327 +0x7be fp=0xc1d07d3898 sp=0xc1d07d3758 pc=0x44e19e
created by github.com/mattermost/mattermost/server/public/plugin.(*hooksRPCClient).OnActivate
	github.com/mattermost/mattermost/server/public/plugin/client_rpc.go:244 +0xab
exit status 2
`

		r := strings.NewReader(input)
		st, err := Parse(r)
		require.NoError(t, err)
		require.Len(t, st, 4)

		var i int
		require.Equal(t, 1, st[i].GoroutineNumber)
		require.Equal(t, "running", st[i].GoroutineState)
		require.Len(t, st[i].StackFrames, 11)
		assert.Equal(t, "reflect.mapiternext", st[i].StackFrames[0].FunctionName)
		assert.Equal(t, "/usr/local/go/src/runtime/map.go", st[i].StackFrames[0].FileName)
		assert.Equal(t, 1532, st[i].StackFrames[0].FileLine)

		i++
		require.Equal(t, 18, st[i].GoroutineNumber)
		require.Equal(t, "runnable", st[i].GoroutineState)
		require.Len(t, st[i].StackFrames, 2)
		assert.Equal(t, "main.main.func1", st[i].StackFrames[0].FunctionName)
		assert.Equal(t, "/home/bschumacher/src/tmp/maps/main.go", st[i].StackFrames[0].FileName)
		assert.Equal(t, 16, st[i].StackFrames[0].FileLine)

		i++
		require.Equal(t, 19, st[i].GoroutineNumber)
		require.Equal(t, "running", st[i].GoroutineState)
		require.Len(t, st[i].StackFrames, 2)
		assert.Equal(t, "main.main.func2", st[i].StackFrames[0].FunctionName)
		assert.Equal(t, "/home/bschumacher/src/tmp/maps/main.go", st[i].StackFrames[0].FileName)
		assert.Equal(t, 23, st[i].StackFrames[0].FileLine)
	})
}
