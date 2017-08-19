package main

import "fmt"

type Logger struct {
	info  bool
	debug bool
}

func (l Logger) Info(msg string) (err error) {
	if l.info {
		_, err = fmt.Printf("INFO: %s\n", msg)
	}
	return err
}

func (l Logger) Debug(msg string) (err error) {
	if l.debug {
		_, err = fmt.Printf("DEBUG: %s\n", msg)
	}
	return err
}

func (l Logger) Error(msg string) (err error) {
	_, err = fmt.Printf("ERROR: %s\n", msg)
	return err
}
