package tools

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func InsertIntoFile(fileA, fileB, fileC, fileD string) error {
	var linesToInsertBefore, linesToInsertAfter []string

	// Read first and last lines of fileB
	fb, err := os.Open(fileB)
	if err != nil {
		return err
	}
	defer fb.Close()
	scannerB := bufio.NewScanner(fb)
	var firstLineB, lastLineB string
	if scannerB.Scan() {
		firstLineB = scannerB.Text()
	}
	for scannerB.Scan() {
		lastLineB = scannerB.Text()
	}

	fbFields := strings.Fields(firstLineB)
	if len(fbFields) < 1 {
		return fmt.Errorf("fileB first line empty")
	}
	firstColumnB, _ := strconv.Atoi(fbFields[0])

	lbFields := strings.Fields(lastLineB)
	if len(lbFields) < 2 {
		return fmt.Errorf("fileB last line empty or insufficient columns")
	}
	lastColumnB, _ := strconv.Atoi(lbFields[1])

	// Process fileA
	fa, err := os.Open(fileA)
	if err != nil {
		return err
	}
	defer fa.Close()
	scannerA := bufio.NewScanner(fa)
	for scannerA.Scan() {
		line := scannerA.Text()
		columns := strings.Fields(line)
		if len(columns) < 2 {
			continue
		}
		firstColumnA, _ := strconv.Atoi(columns[0])
		secondColumnA, _ := strconv.Atoi(columns[1])

		if secondColumnA > firstColumnB && firstColumnA < firstColumnB {
			columns[1] = strconv.Itoa(firstColumnB)
		}
		
		modifiedLine := strings.Join(columns, " ")
		if firstColumnA < firstColumnB {
			linesToInsertBefore = append(linesToInsertBefore, modifiedLine)
		} else if firstColumnA > lastColumnB {
			linesToInsertAfter = append(linesToInsertAfter, modifiedLine)
		}
	}

	// Read fileC
	var linesC []string
	fc, err := os.Open(fileC)
	if err != nil {
		return err
	}
	defer fc.Close()
	scannerC := bufio.NewScanner(fc)
	for scannerC.Scan() {
		linesC = append(linesC, scannerC.Text())
	}

	// Combine and write fileD
	fd, err := os.Create(fileD)
	if err != nil {
		return err
	}
	defer fd.Close()
	writer := bufio.NewWriter(fd)

	for _, line := range linesToInsertBefore {
		writer.WriteString(strings.Join(strings.Fields(line), "\t") + "\n")
	}
	for _, line := range linesC {
		writer.WriteString(strings.Join(strings.Fields(line), "\t") + "\n")
	}
	for _, line := range linesToInsertAfter {
		writer.WriteString(strings.Join(strings.Fields(line), "\t") + "\n")
	}
	return writer.Flush()
}

