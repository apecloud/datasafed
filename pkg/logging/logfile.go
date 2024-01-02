package logging

import (
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/apecloud/datasafed/pkg/util"
)

const logsDirMode = 0o700

var logLevels = []string{"debug", "info", "warning", "error"}

type loggingFlags struct {
	logFile               string
	logDir                string
	logDirMaxFiles        int
	logDirMaxAge          time.Duration
	logDirMaxTotalSizeMB  float64
	logFileMaxSegmentSize int
	logLevel              string
	fileLogLevel          string
	fileLogLocalTimezone  bool
	jsonLogFile           bool
	jsonLogConsole        bool
	forceColor            bool
	disableColor          bool
	consoleLog            bool
	consoleLogTimestamps  bool
}

func (c *loggingFlags) setup(cmd *cobra.Command) {
	f := cmd.PersistentFlags()
	f.StringVar(&c.logFile, "log-file", "", "Override log file.")
	f.StringVar(&c.logDir, "log-dir", "", "Directory where log files should be written.")
	f.IntVar(&c.logDirMaxFiles, "log-dir-max-files", 100, "Maximum number of log files to retain")
	f.DurationVar(&c.logDirMaxAge, "log-dir-max-age", 720*time.Hour, "Maximum age of log files to retain")
	f.Float64Var(&c.logDirMaxTotalSizeMB, "log-dir-max-total-size-mb", 1000, "Maximum total size of log files to retain")
	f.IntVar(&c.logFileMaxSegmentSize, "max-log-file-segment-size", 50000000, "Maximum size of a single log file segment")
	f.Var(util.NewEnumVar(logLevels, &c.logLevel).Default("info"), "log-level", "Console log level")
	f.BoolVar(&c.jsonLogConsole, "json-log-console", false, "JSON log file")
	f.BoolVar(&c.jsonLogFile, "json-log-file", false, "JSON log file")
	f.Var(util.NewEnumVar(logLevels, &c.fileLogLevel).Default("debug"), "file-log-level", "File log level")
	f.BoolVar(&c.fileLogLocalTimezone, "file-log-local-tz", false, "When logging to a file, use local timezone")
	f.BoolVar(&c.forceColor, "force-color", false, "Force color output")
	f.BoolVar(&c.disableColor, "disable-color", false, "Disable color output")
	f.BoolVar(&c.consoleLog, "console-log", false, "Enable console log")
	f.BoolVar(&c.consoleLogTimestamps, "console-timestamps", true, "Log timestamps to stderr.")

	old := cmd.PersistentPreRunE
	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := c.initialize(cmd); err != nil {
			return err
		}
		if old != nil {
			return old(cmd, args)
		}
		return nil
	}
}

// Attach attaches logging flags to the provided application.
func Attach(rootCmd *cobra.Command) {
	lf := &loggingFlags{}
	lf.setup(rootCmd)
}

var (
	log = Module("datasafed")

	DefaultLoggerFactory LoggerFactory
)

const (
	logFileNamePrefix = "datasafed-"
	logFileNameSuffix = ".log"
)

// initialize is invoked as part of command execution to create log file just before it's needed.
func (c *loggingFlags) initialize(cmd *cobra.Command) error {
	now := time.Now()
	if c.fileLogLocalTimezone {
		now = now.Local()
	} else {
		now = now.UTC()
	}

	var cores []zapcore.Core
	if c.consoleLog {
		cores = append(cores, c.setupConsoleCore(cmd.ErrOrStderr()))
	}
	if c.logDir != "" {
		suffix := strings.ReplaceAll(cmd.Name(), " ", "-")
		cores = append(cores, c.setupLogFileCore(now, suffix))
	}
	if len(cores) == 0 {
		return nil
	}

	rootLogger := zap.New(zapcore.NewTee(cores...), zap.WithClock(Clock()))

	DefaultLoggerFactory = func(module string) Logger {
		return rootLogger.Named(module).Sugar()
	}

	if c.forceColor {
		color.NoColor = false
	}

	if c.disableColor {
		color.NoColor = true
	}

	return nil
}

func (c *loggingFlags) setupConsoleCore(output io.Writer) zapcore.Core {
	ec := zapcore.EncoderConfig{
		LevelKey:         "l",
		MessageKey:       "m",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeTime:       zapcore.RFC3339NanoTimeEncoder,
		EncodeDuration:   zapcore.StringDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		ConsoleSeparator: " ",
	}

	timeFormat := PreciseLayout

	if c.consoleLogTimestamps {
		ec.TimeKey = "t"

		if c.jsonLogConsole {
			ec.EncodeTime = zapcore.RFC3339NanoTimeEncoder
		} else {
			// always log local timestamps to the console, not UTC
			timeFormat = "15:04:05.000"
			ec.EncodeTime = TimezoneAdjust(zapcore.TimeEncoderOfLayout(timeFormat), true)
		}
	} else {
		timeFormat = ""
	}

	stec := StdConsoleEncoderConfig{
		TimeLayout: timeFormat,
		LocalTime:  true,
	}

	if c.jsonLogConsole {
		ec.EncodeLevel = zapcore.CapitalLevelEncoder

		ec.NameKey = "n"
		ec.EncodeName = zapcore.FullNameEncoder
	} else {
		stec.EmitLogLevel = true
		stec.DoNotEmitInfoLevel = true
		stec.ColoredLogLevel = !c.disableColor
	}

	return zapcore.NewCore(
		c.jsonOrConsoleEncoder(stec, ec, c.jsonLogConsole),
		zapcore.AddSync(output),
		logLevelFromFlag(c.logLevel),
	)
}

func (c *loggingFlags) setupLogFileBasedLogger(now time.Time, subdir, suffix, logFileOverride string, maxFiles int, maxSizeMB float64, maxAge time.Duration) zapcore.WriteSyncer {
	var logFileName, symlinkName string

	if logFileOverride != "" {
		var err error

		logFileName, err = filepath.Abs(logFileOverride)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Unable to resolve logs path", err)
		}
	}

	if logFileName == "" {
		logBaseName := fmt.Sprintf("%v%v-%v-%v%v", logFileNamePrefix, now.Format("20060102-150405"), os.Getpid(), suffix, logFileNameSuffix)
		logFileName = filepath.Join(c.logDir, subdir, logBaseName)
		symlinkName = "latest.log"
	}

	logDir := filepath.Dir(logFileName)
	logFileBaseName := filepath.Base(logFileName)

	if err := os.MkdirAll(logDir, logsDirMode); err != nil {
		fmt.Fprintln(os.Stderr, "Unable to create logs directory:", err)
	}

	sweepLogWG := &sync.WaitGroup{}
	doSweep := func() {}

	// do not scrub directory if custom log file has been provided.
	if logFileOverride == "" && shouldSweepLog(maxFiles, maxAge) {
		doSweep = func() {
			sweepLogDir(context.TODO(), logDir, maxFiles, maxSizeMB, maxAge)
		}
	}

	odf := &onDemandFile{
		logDir:          logDir,
		logFileBaseName: logFileBaseName,
		symlinkName:     symlinkName,
		maxSegmentSize:  c.logFileMaxSegmentSize,
		startSweep: func() {
			sweepLogWG.Add(1)

			go func() {
				defer sweepLogWG.Done()

				doSweep()
			}()
		},
	}

	// old behavior: start log sweep in parallel to program but don't wait at the end.
	odf.startSweep()

	return odf
}

func (c *loggingFlags) setupLogFileCore(now time.Time, suffix string) zapcore.Core {
	return zapcore.NewCore(
		c.jsonOrConsoleEncoder(
			StdConsoleEncoderConfig{
				TimeLayout:     PreciseLayout,
				LocalTime:      c.fileLogLocalTimezone,
				EmitLogLevel:   true,
				EmitLoggerName: true,
			},
			zapcore.EncoderConfig{
				TimeKey:          "t",
				MessageKey:       "m",
				NameKey:          "n",
				LevelKey:         "l",
				EncodeName:       zapcore.FullNameEncoder,
				EncodeLevel:      zapcore.CapitalLevelEncoder,
				EncodeTime:       TimezoneAdjust(PreciseTimeEncoder(), c.fileLogLocalTimezone),
				EncodeDuration:   zapcore.StringDurationEncoder,
				ConsoleSeparator: " ",
			},
			c.jsonLogFile),
		c.setupLogFileBasedLogger(now, "cli-logs", suffix, c.logFile, c.logDirMaxFiles, c.logDirMaxTotalSizeMB, c.logDirMaxAge),
		logLevelFromFlag(c.fileLogLevel),
	)
}

//nolint:gocritic
func (c *loggingFlags) jsonOrConsoleEncoder(ec StdConsoleEncoderConfig, jc zapcore.EncoderConfig, isJSON bool) zapcore.Encoder {
	if isJSON {
		return zapcore.NewJSONEncoder(jc)
	}

	return NewStdConsoleEncoder(ec)
}

func shouldSweepLog(maxFiles int, maxAge time.Duration) bool {
	return maxFiles > 0 || maxAge > 0
}

func sweepLogDir(ctx context.Context, dirname string, maxCount int, maxSizeMB float64, maxAge time.Duration) {
	var timeCutoff time.Time
	if maxAge > 0 {
		timeCutoff = time.Now().Add(-maxAge)
	}

	if maxCount == 0 {
		maxCount = math.MaxInt32
	}

	maxTotalSizeBytes := int64(maxSizeMB * 1e6)

	entries, err := os.ReadDir(dirname)
	if err != nil {
		log(ctx).Errorf("unable to read log directory: %v", err)
		return
	}

	fileInfos := make([]os.FileInfo, 0, len(entries))

	for _, e := range entries {
		info, err2 := e.Info()
		if os.IsNotExist(err2) {
			// we lost the race, the file was deleted since it was listed.
			continue
		}

		if err2 != nil {
			log(ctx).Errorf("unable to read file info: %v", err2)
			return
		}

		fileInfos = append(fileInfos, info)
	}

	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].ModTime().After(fileInfos[j].ModTime())
	})

	cnt := 0
	totalSize := int64(0)

	for _, fi := range fileInfos {
		if !strings.HasPrefix(fi.Name(), logFileNamePrefix) {
			continue
		}

		if !strings.HasSuffix(fi.Name(), logFileNameSuffix) {
			continue
		}

		cnt++

		totalSize += fi.Size()

		if cnt > maxCount || totalSize > maxTotalSizeBytes || fi.ModTime().Before(timeCutoff) {
			if err = os.Remove(filepath.Join(dirname, fi.Name())); err != nil && !os.IsNotExist(err) {
				log(ctx).Errorf("unable to remove log file: %v", err)
			}
		}
	}
}

func logLevelFromFlag(levelString string) zapcore.LevelEnabler {
	switch levelString {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warning":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	default:
		return zap.FatalLevel
	}
}

type onDemandFile struct {
	// +checklocks:mu
	segmentCounter int // number of segments written

	// +checklocks:mu
	currentSegmentSize int // number of bytes written to current segment

	// +checklocks:mu
	maxSegmentSize int

	// +checklocks:mu
	currentSegmentFilename string

	// +checklocks:mu
	logDir string

	// +checklocks:mu
	logFileBaseName string

	// +checklocks:mu
	symlinkName string

	startSweep func()

	mu sync.Mutex
	f  *os.File
}

func (w *onDemandFile) Sync() error {
	if w.f == nil {
		return nil
	}

	//nolint:wrapcheck
	return w.f.Sync()
}

func (w *onDemandFile) closeSegmentAndSweepLocked() {
	if w.f != nil {
		if err := w.f.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: unable to close log segment: %v", err)
		}

		w.f = nil
	}

	w.startSweep()
}

func (w *onDemandFile) Write(b []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// close current file if we'd overflow on next write.
	if w.f != nil && w.currentSegmentSize+len(b) > w.maxSegmentSize {
		w.closeSegmentAndSweepLocked()
	}

	// open file if we don't have it yet
	if w.f == nil {
		var baseName, ext string

		p := strings.LastIndex(w.logFileBaseName, ".")
		if p < 0 {
			ext = ""
			baseName = w.logFileBaseName
		} else {
			ext = w.logFileBaseName[p:]
			baseName = w.logFileBaseName[0:p]
		}

		w.currentSegmentFilename = fmt.Sprintf("%s.%d%s", baseName, w.segmentCounter, ext)
		w.segmentCounter++
		w.currentSegmentSize = 0

		lf := filepath.Join(w.logDir, w.currentSegmentFilename)

		f, err := os.Create(lf) //nolint:gosec
		if err != nil {
			return 0, fmt.Errorf("unable to open log file, %w", err)
		}

		w.f = f

		if w.symlinkName != "" {
			symlink := filepath.Join(w.logDir, w.symlinkName)
			_ = os.Remove(symlink)                            // best-effort remove
			_ = os.Symlink(w.currentSegmentFilename, symlink) // best-effort symlink
		}
	}

	n, err := w.f.Write(b)
	w.currentSegmentSize += n

	//nolint:wrapcheck
	return n, err
}
