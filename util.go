package util

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func LoadEnvFile(path string) error {
	f, err := os.Open(path)

	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)

	output := make(map[string]string)

	for scanner.Scan() {
		text := scanner.Text()
		line := strings.Split(text, "=")

		if len(line) != 2 {
			continue
		}

		key := line[0]
		val := line[1]

		os.Setenv(key, val)

		output[key] = val
	}

	return nil
}

func Contains[T comparable](haystack []T, needle T) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}

	return false
}
