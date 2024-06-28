package log

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

const backupTimeFormat = "2006-01-02_15"

type TimeTicker struct {
	stop chan struct{}
	C    <-chan time.Time
}

func NewTimeTicker(rotateHours uint) *TimeTicker {
	ch := make(chan time.Time)
	tt := TimeTicker{
		stop: make(chan struct{}),
		C:    ch,
	}

	if rotateHours > 0 {
		tt.startTicker(ch, rotateHours)
	}

	return &tt
}

func (tt *TimeTicker) Stop() {
	tt.stop <- struct{}{}
}

func (tt *TimeTicker) startTicker(ch chan time.Time, rotateHours uint) {
	go func() {
		nextRotationHour := getNextRotationHour(time.Now(), rotateHours)
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case t := <-ticker.C:
				if t.Hour() == nextRotationHour {
					ch <- t
					nextRotationHour = getNextRotationHour(time.Now(), rotateHours)
				}
			case <-tt.stop:
				return
			}
		}
	}()
}

func getNextRotationHour(now time.Time, delta uint) int {
	return now.Add(time.Hour * time.Duration(delta)).Hour()
}

type FileHandlerOptions struct {
	Level       slog.Level
	FilePath    string
	Limit       int64
	MaxBackups  int
	RotateHours uint
}

type AsyncFileWriter struct {
	filePath    string
	fd          *os.File
	wg          sync.WaitGroup
	started     int32
	buf         chan []byte
	stop        chan struct{}
	timeTicker  *TimeTicker
	rotateHours uint
	maxBackups  int
}

func (w *AsyncFileWriter) initLogFile() error {
	var (
		fd  *os.File
		err error
	)

	realFilePath := w.timeFilePath(w.filePath)
	fd, err = os.OpenFile(realFilePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	w.fd = fd
	symLinkFilePath := filepath.Join(w.filePath, "log.log")
	_, err = os.Lstat(symLinkFilePath)
	if err == nil || os.IsExist(err) {
		err = os.Remove(symLinkFilePath)
		if err != nil {
			return err
		}
	}
	_ = os.Symlink(realFilePath, symLinkFilePath)
	return nil
}

func (w *AsyncFileWriter) Start() error {
	if !atomic.CompareAndSwapInt32(&w.started, 0, 1) {
		return errors.New("logger has already been started")
	}

	err := w.initLogFile()
	if err != nil {
		return err
	}

	w.wg.Add(1)
	go func() {
		defer func() {
			atomic.StoreInt32(&w.started, 0)

			w.flushBuffer()
			w.flushAndClose()

			w.wg.Done()
		}()

		for {
			select {
			case msg, ok := <-w.buf:
				if !ok {
					fmt.Fprintln(os.Stderr, "buf channel has been closed.")
					return
				}
				w.SyncWrite(msg)
			case <-w.stop:
				return
			}
		}
	}()
	return nil
}

func (w *AsyncFileWriter) flushBuffer() {
	for {
		select {
		case msg := <-w.buf:
			w.SyncWrite(msg)
		default:
			return
		}
	}
}

func (w *AsyncFileWriter) SyncWrite(msg []byte) {
	w.rotateFile()
	if w.fd != nil {
		w.fd.Write(msg)
	}
}

func (w *AsyncFileWriter) rotateFile() {
	select {
	case <-w.timeTicker.C:
		if err := w.flushAndClose(); err != nil {
			fmt.Fprintf(os.Stderr, "flush and close file error. err=%s", err)
		}
		if err := w.initLogFile(); err != nil {
			fmt.Fprintf(os.Stderr, "init log file error. err=%s", err)
		}
		if err := w.removeExpiredFile(); err != nil {
			fmt.Fprintf(os.Stderr, "remove expired file error. err=%s", err)
		}
	default:
	}
}

func (w *AsyncFileWriter) Stop() {
	w.stop <- struct{}{}
	w.wg.Wait()

	w.timeTicker.Stop()
}

func (w *AsyncFileWriter) Write(msg []byte) (n int, err error) {
	buf := make([]byte, len(msg))
	copy(buf, msg)

	select {
	case w.buf <- buf:
	default:

	}
	return 0, nil
}

func (w *AsyncFileWriter) Flush() error {
	if w.fd == nil {
		return nil
	}
	return w.fd.Sync()
}

func (w *AsyncFileWriter) flushAndClose() error {
	if w.fd == nil {
		return nil
	}

	err := w.fd.Sync()
	if err != nil {
		return err
	}

	return w.fd.Close()
}

func (w *AsyncFileWriter) timeFilePath(path string) string {
	return path + "." + time.Now().Format(backupTimeFormat)
}

func (w *AsyncFileWriter) getExpiredFile(filePath string, maxBackups int, rotateHours uint) string {
	if rotateHours > 0 {
		maxBackups = int(rotateHours) * maxBackups
	}

	return filePath + "." + time.Now().Add(-time.Hour*time.Duration(maxBackups)).Format(backupTimeFormat)
}

func (w *AsyncFileWriter) removeExpiredFile() error {
	if w.maxBackups == 0 {
		return nil
	}

	oldFilepath := w.getExpiredFile(w.filePath, w.maxBackups, w.rotateHours)

	_, err := os.Stat(oldFilepath)
	if os.IsNotExist(err) {
		return nil
	}
	errRemove := os.Remove(oldFilepath)
	if err != nil {
		return errRemove
	}
	return err
}

func NewAsyncFileWriter(opts *FileHandlerOptions) *AsyncFileWriter {
	if opts == nil {
		opts = &FileHandlerOptions{}
	}
	if opts.FilePath == "" {
		opts.FilePath = "log"
	}
	if _, err := os.Stat(filepath.Dir(opts.FilePath)); os.IsNotExist(err) {
		err := os.MkdirAll(filepath.Dir(opts.FilePath), 0755)
		if err != nil {
			Fatal("could not create directory", "directory", opts.FilePath, "err", err)
		}
	}
	absFilePath, err := filepath.Abs(opts.FilePath)
	if err != nil {
		Fatal("could not get abs directory", "err", err)
	}
	writer := &AsyncFileWriter{
		filePath:    absFilePath,
		buf:         make(chan []byte, opts.Limit),
		stop:        make(chan struct{}),
		rotateHours: opts.RotateHours,
		maxBackups:  opts.MaxBackups,
		timeTicker:  NewTimeTicker(opts.RotateHours),
	}
	if err = writer.Start(); err != nil {
		Fatal("could not start async file handler", "err", err)
	}
	return writer
}

func NewTextHandler(w io.Writer, opts *slog.HandlerOptions) *slog.TextHandler {
	return slog.NewTextHandler(w, opts)
}
