package tools

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ProcessTiaozhen runs the logic similar to tiaozhen.py
func ProcessTiaozhen(inputFile, outputFile, param1, param2, param3 string) error {
	in, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer in.Close()

	var lines []string
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) != "Source_Strain\tSource_st\tSource_end\tAnnotation\tHom_strain" {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if len(lines) < 1 {
		return fmt.Errorf("input file content does not meet expected format (too short)")
	}

	out, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer out.Close()
	writer := bufio.NewWriter(out)

	// Write header back
	_, err = writer.WriteString("Source_Strain\tSource_st\tSource_end\tAnnotation\tHom_strain\n")
	if err != nil {
		return err
	}

	for _, line := range lines {
		parts := strings.Split(strings.TrimSpace(line), "\t")
		if len(parts) < 5 {
			continue
		}

		isDigit := true
		for _, char := range parts[0] {
			if char < '0' || char > '9' {
				isDigit = false
				break
			}
		}

		if isDigit && parts[0] != param1 && parts[0] != param3 {
			newLine := fmt.Sprintf("%s\t%s\t%s\t%s\t%s_added\n", param1, parts[0], parts[1], param2, param2)
			writer.WriteString(newLine)
		} else {
			writer.WriteString(strings.Join(parts, "\t") + "\n")
		}
	}

	return writer.Flush()
}
