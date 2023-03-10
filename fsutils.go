package main

import (
	"log"
	"os"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func IsDir(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}

	return fi.IsDir()
}

func CreateDir(path string) {
	err := os.MkdirAll(path, 0770)
	if err != nil {
		log.Fatal(err)
	}
}
