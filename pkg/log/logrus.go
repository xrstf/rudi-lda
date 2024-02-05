package log

import (
	"log"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

var directory = ""

func SetDirectory(dirname string) error {
	if err := os.MkdirAll(dirname, 0700); err != nil {
		return err
	}

	directory = dirname

	return nil
}

func New(filename string) *logrus.Logger {
	f, err := os.OpenFile(filepath.Join(directory, filename), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
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
