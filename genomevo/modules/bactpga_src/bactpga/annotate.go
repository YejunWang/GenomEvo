package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func runAnnotate(pgmapmatFileName, agFileName, mode, strain, mutbestDir, seqDir string) {
	// Parse 26_PG.txt file
	pgmapFile, openErr1 := os.Open(pgmapmatFileName)
	if openErr1 != nil {
		fmt.Printf("Error happened in opening %s!\n", pgmapmatFileName)
		return
	}
	defer pgmapFile.Close()

	pg := make(map[string]string)
	pgmapScanner := bufio.NewScanner(pgmapFile)
	for pgmapScanner.Scan() {
		pgmap := strings.TrimSpace(pgmapScanner.Text())
		if len(pgmap) > 0 {
			pgmapArray := strings.Split(pgmap, "\t")
			if len(pgmapArray) > 2 {
				for i := 2; i < len(pgmapArray)-1; i++ {
					if pgmapArray[i] != "-" {
						pg[pgmapArray[i]] = pgmapArray[0]
					}
				}
			}
		}
	}
	if err := pgmapScanner.Err(); err != nil {
		fmt.Printf("Error reading %s: %v\n", pgmapmatFileName, err)
		return
	}

	var flag int
	if strings.Contains(mode, "PGAG") {
		flag = 2
	} else if strings.Contains(mode, "RAST") {
		flag = 1
	} else {
		fmt.Println("Mode must be PGAG or RAST")
		return
	}

	annMap := make(map[string]string)

	// Determine mutual best hit files
	var mutBestFiles []string
	if seqDir != "" && seqDir != "." {
		// Legacy compatibility mapping `count` logic based on sequences dir
		entries, err := os.ReadDir(seqDir)
		if err != nil {
			fmt.Printf("Error reading seqDir %s: %v\n", seqDir, err)
			return
		}
		count := 0
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".faa") {
				count++
			}
		}
		for i := 1; i <= count; i++ {
			mutbestFileName := filepath.Join(mutbestDir, fmt.Sprintf("%s.vs.%d.mutbest.filt.txt", strain, i))
			mutBestFiles = append(mutBestFiles, mutbestFileName)
		}
	} else {
		// Just glob all mutbest files in mutbestDir
		pattern := filepath.Join(mutbestDir, strain+".vs.*.mutbest.filt.txt")
		var err error
		mutBestFiles, err = filepath.Glob(pattern)
		if err != nil {
			fmt.Printf("Error matching .mutbest.filt.txt files: %v\n", err)
			return
		}
	}

	// Read mutbest files to build annotation map
	for _, mutbestFileName := range mutBestFiles {
		parseMutBest(mutbestFileName, annMap, pg)
	}

	// Parse main table and annotate
	agFile, openErr3 := os.Open(agFileName)
	if openErr3 != nil {
		fmt.Printf("Error happened in opening %s!\n", agFileName)
		return
	}
	defer agFile.Close()

	agScanner := bufio.NewScanner(agFile)
	for agScanner.Scan() {
		ag := strings.TrimSpace(agScanner.Text())
		if len(ag) > 0 {
			agArray := strings.Split(ag, "\t")
			
			// Output column up to flag
			for i := 0; i <= flag && i < len(agArray); i++ {
				fmt.Printf("%s\t", agArray[i])
			}
			
			// Inject annotation
			if flag < len(agArray) && len(annMap[agArray[flag]]) > 0 {
				fmt.Printf("%s", annMap[agArray[flag]])
			}
			
			// Output remaining columns
			for i := flag + 1; i < len(agArray); i++ {
				fmt.Printf("\t%s", agArray[i])
			}
			fmt.Printf("\n")
		}
	}
}

func parseMutBest(mutbestFileName string, annMap, pg map[string]string) {
	mutbestFile, err := os.Open(mutbestFileName)
	if err != nil {
		// Ignore gracefully as it might be missing or user typo
		return
	}
	defer mutbestFile.Close()

	scanner := bufio.NewScanner(mutbestFile)
	for scanner.Scan() {
		mutbest := strings.TrimSpace(scanner.Text())
		if len(mutbest) > 0 {
			mutbestArray := strings.Split(mutbest, "\t")
			if len(mutbestArray) >= 2 {
				query := mutbestArray[0]
				subject := mutbestArray[1]
				if annMap[query] == "" {
					if pgUID, ok := pg[subject]; ok && pgUID != "" {
						annMap[query] = pgUID
					}
				}
			}
		}
	}
}
