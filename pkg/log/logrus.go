package log

import (
	"log"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"go.xrstf.de/rudi-lda/pkg/fs"
)

var directory = ""

func SetDirectory(dirname string) error {
	if err := os.MkdirAll(dirname, fs.DirectoryPermissions); err != nil {
		return err
	}

	directory = dirname

	return nil
}

func New(filename string) *logrus.Logger {
	f, err := os.OpenFile(filepath.Join(directory, filename), os.O_RDWR|os.O_CREATE|os.O_APPEND, fs.FilePermissions)
	if err != nil {
		log.Fatalf("Cannot open log file %q: %v", filename, err)
	}
	// f stays open until the program ends

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{
		DisableHTMLEscape: true,
	})
	logger.SetOutput(f)

	return logger
}
