package templates

import "errors"

// LanguageLookup returns the Docker image name for a given programming language.
func LanguageLookup(lang string) (string, error) {
	languages := map[string]string{
		"go":     "golang:1.24",
		"python": "python:3.9",
		"node":   "node:14",
		"java":   "openjdk:11",
		"ruby":   "ruby:2.7",
		"php":    "php:8.0",
		"rust":   "rust:1.56",
	}

	if file, exists := languages[lang]; exists {
		return file, nil
	}
	return "", errors.New("unsupported language: " + lang)
}
	