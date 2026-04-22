package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func loadFromFile(path string) []string {

	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("[WARN] cannot read channel file: %v\n", err)
		return nil
	}
	defer file.Close()

	var result []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		result = append(result, line)
	}

	return result
}
