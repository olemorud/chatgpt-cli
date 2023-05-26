package util

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func ReadEnvFile(path string) (map[string]string, error) {
	f, err := os.Open(path)

	if err != nil {
		fmt.Println("failed to open file: ", err)
		return nil, fmt.Errorf("failed to open file: %v", err)
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

		output[key] = val
	}

	return output, nil
}

func Contains[T comparable](haystack []T, needle T) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}

	return false
}
