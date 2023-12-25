package kopia

import (
	"github.com/kopia/kopia/snapshot/snapshotfs"

	"github.com/apecloud/datasafed/pkg/logging"
)

// LoggingUploadProgress is an implementation of UploadProgress.
type LoggingUploadProgress struct {
	logger logging.Logger
}

// UploadStarted implements UploadProgress.
func (p *LoggingUploadProgress) UploadStarted() {
	p.logger.Infof("[PROGRESS] Upload started")
}

// EstimatedDataSize implements UploadProgress.
func (p *LoggingUploadProgress) EstimatedDataSize(fileCount int, totalBytes int64) {
	p.logger.Infof("[PROGRESS] Estimated data size: %d files, %d bytes", fileCount, totalBytes)
}

// UploadFinished implements UploadProgress.
func (p *LoggingUploadProgress) UploadFinished() {
	p.logger.Infof("[PROGRESS] Upload finished")
}

// HashedBytes implements UploadProgress.
func (p *LoggingUploadProgress) HashedBytes(numBytes int64) {
	p.logger.Infof("[PROGRESS] Hashed %d bytes", numBytes)
}

// ExcludedFile implements UploadProgress.
func (p *LoggingUploadProgress) ExcludedFile(fname string, numBytes int64) {
	p.logger.Infof("[PROGRESS] Excluded %s (%d bytes)", fname, numBytes)
}

// ExcludedDir implements UploadProgress.
func (p *LoggingUploadProgress) ExcludedDir(dirname string) {
	p.logger.Infof("[PROGRESS] Excluded %s", dirname)
}

// CachedFile implements UploadProgress.
func (p *LoggingUploadProgress) CachedFile(fname string, numBytes int64) {
	p.logger.Infof("[PROGRESS] Cached %s (%d bytes)", fname, numBytes)
}

// UploadedBytes implements UploadProgress.
func (p *LoggingUploadProgress) UploadedBytes(numBytes int64) {
	p.logger.Infof("[PROGRESS] Uploaded %d bytes", numBytes)
}

// HashingFile implements UploadProgress.
func (p *LoggingUploadProgress) HashingFile(fname string) {
	p.logger.Infof("[PROGRESS] Hashing %s", fname)
}

// FinishedHashingFile implements UploadProgress.
func (p *LoggingUploadProgress) FinishedHashingFile(fname string, numBytes int64) {
	p.logger.Infof("[PROGRESS] Finished hashing %s (%d bytes)", fname, numBytes)
}

// FinishedFile implements UploadProgress.
func (p *LoggingUploadProgress) FinishedFile(fname string, err error) {
	p.logger.Infof("[PROGRESS] Finished %s (%v)", fname, err)
}

// StartedDirectory implements UploadProgress.
func (p *LoggingUploadProgress) StartedDirectory(dirname string) {
	p.logger.Infof("[PROGRESS] Started %s", dirname)
}

// FinishedDirectory implements UploadProgress.
func (p *LoggingUploadProgress) FinishedDirectory(dirname string) {
	p.logger.Infof("[PROGRESS] Finished %s", dirname)
}

// Error implements UploadProgress.
func (p *LoggingUploadProgress) Error(path string, err error, isIgnored bool) {
	p.logger.Infof("[PROGRESS] Error %s (%v, ignored=%v)", path, err, isIgnored)
}

var _ snapshotfs.UploadProgress = (*LoggingUploadProgress)(nil)

func newLoggingUploadProgress(logger logging.Logger) *LoggingUploadProgress {
	return &LoggingUploadProgress{logger: logger}
}
