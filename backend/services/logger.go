package services

import (
	"log"
)

func Info(msg string) {
	log.Printf("\033[34m[INFO]\033[0m %s\n", msg)
}

func Success(msg string) {
	log.Printf("\033[32m[SUCCESS]\033[0m %s\n", msg)
}

func Warn(msg string) {
	log.Printf("\033[33m[WARN]\033[0m %s\n", msg)
}

func Error(msg string) {
	log.Printf("\033[31m[ERROR]\033[0m %s\n", msg)
}
