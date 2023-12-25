package logging_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/apecloud/datasafed/pkg/logging"
)

// Printf returns a logger that uses given printf-style function to print log output.
func Printf(printf func(msg string, args ...interface{}), prefix string) *zap.SugaredLogger {
	return PrintfLevel(printf, prefix, zapcore.DebugLevel)
}

// PrintfLevel returns a logger that uses given printf-style function to print log output for logs of a given level or above.
func PrintfLevel(printf func(msg string, args ...interface{}), prefix string, level zapcore.Level) *zap.SugaredLogger {
	writer := printfWriter{printf, prefix}

	return zap.New(
		zapcore.NewCore(
			logging.NewStdConsoleEncoder(logging.StdConsoleEncoderConfig{}),
			writer,
			level,
		),
	).Sugar()
}

// PrintfFactory returns LoggerForModuleFunc that uses given printf-style function to print log output.
func PrintfFactory(printf func(msg string, args ...interface{})) logging.LoggerFactory {
	return func(module string) *zap.SugaredLogger {
		return Printf(printf, "["+module+"] ")
	}
}

type printfWriter struct {
	printf func(msg string, args ...interface{})
	prefix string
}

func (w printfWriter) Write(p []byte) (int, error) {
	n := len(p)

	w.printf("%s%s", w.prefix, bytes.TrimRight(p, "\n"))

	return n, nil
}

func (w printfWriter) Sync() error {
	return nil
}

func TestBroadcast(t *testing.T) {
	var lines []string

	l0 := Printf(func(msg string, args ...interface{}) {
		lines = append(lines, fmt.Sprintf(msg, args...))
	}, "[first] ")

	l1 := Printf(func(msg string, args ...interface{}) {
		lines = append(lines, fmt.Sprintf(msg, args...))
	}, "[second] ")

	l := logging.Broadcast(l0, l1)
	l.Debugf("A")
	l.Debugw("S", "b", 123)
	l.Infof("B")
	l.Errorf("C")
	l.Warnf("W")

	require.Equal(t, []string{
		"[first] A",
		"[second] A",
		"[first] S\t{\"b\":123}",
		"[second] S\t{\"b\":123}",
		"[first] B",
		"[second] B",
		"[first] C",
		"[second] C",
		"[first] W",
		"[second] W",
	}, lines)
}

func TestWriter(t *testing.T) {
	var buf bytes.Buffer

	l := logging.ToWriter(&buf)("module1")
	l.Debugf("A")
	l.Debugw("S", "b", 123)
	l.Infof("B")
	l.Errorf("C")
	l.Warnf("W")

	require.Equal(t, "A\nS\t{\"b\":123}\nB\nC\nW\n", buf.String())
}

func TestNullWriterModule(t *testing.T) {
	l := logging.Module("mod1")(context.Background())

	l.Debugf("A")
	l.Debugw("S", "b", 123)
	l.Infof("B")
	l.Errorf("C")
	l.Warnf("W")
}

func TestNonNullWriterModule(t *testing.T) {
	var buf bytes.Buffer

	ctx := logging.WithLogger(context.Background(), logging.ToWriter(&buf))
	l := logging.Module("mod1")(ctx)

	l.Debugf("A")
	l.Debugw("S", "b", 123)
	l.Infof("B")
	l.Errorf("C")
	l.Warnf("W")

	require.Equal(t, "A\nS\t{\"b\":123}\nB\nC\nW\n", buf.String())
}

func TestWithAdditionalLogger(t *testing.T) {
	var buf, buf2 bytes.Buffer

	ctx := logging.WithLogger(context.Background(), logging.ToWriter(&buf))
	ctx = logging.WithAdditionalLogger(ctx, logging.ToWriter(&buf2))
	l := logging.Module("mod1")(ctx)

	l.Debugf("A")
	l.Debugw("S", "b", 123)
	l.Infof("B")
	l.Errorf("C")
	l.Warnf("W")

	require.Equal(t, "A\nS\t{\"b\":123}\nB\nC\nW\n", buf.String())
	require.Equal(t, "A\nS\t{\"b\":123}\nB\nC\nW\n", buf2.String())
}

func BenchmarkLogger(b *testing.B) {
	mod1 := logging.Module("mod1")
	ctx := logging.WithLogger(context.Background(), PrintfFactory(b.Logf))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mod1(ctx)
	}
}
