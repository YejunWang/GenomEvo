package main

import (
	"bactag/internal/pipeline"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

func main() {
	if runSubcommand() {
		// Handled entirely by internal standalone router (like busybox)
		return
	}

	var (
		threads      int
		treeFilePath string
		idFilePath   string
		geneDir      string
	)

	flag.IntVar(&threads, "t", 4, "Number of threads to use")
	flag.StringVar(&treeFilePath, "tree", "input/tree_file", "Path to directory containing tree file")
	flag.StringVar(&idFilePath, "id", "codes/bactID.txt", "Path to the ID file")
	flag.StringVar(&geneDir, "gene", "input/gene", "Path to directory containing input genome FASTA files")
	flag.Parse()

	// Check if progressiveMauve is installed
	if _, err := exec.LookPath("progressiveMauve"); err != nil {
		log.Fatalf("Error: Dependency 'progressiveMauve' not found. Please install it before running BactAG.\n" +
			"For Ubuntu/Debian, you can try: sudo apt install mauve-aligner")
	}

	log.Printf("Starting BactAG with %d threads", threads)

	// Pre-Execution Lifecycle (from BACT_AG.sh)
	log.Println("Initializing working directories and clearing old cache...")
	os.RemoveAll("log")
	os.RemoveAll("output")
	os.RemoveAll("temp")
	os.Remove(idFilePath)

	// Setup timestamp and ID file
	currTime := time.Now().Format("2006-01-02 15:04:05")
	os.MkdirAll(filepath.Dir(idFilePath), 0755)
	idContent := fmt.Sprintf("# This is BactID  %s\n", currTime)
	if err := ioutil.WriteFile(idFilePath, []byte(idContent), 0644); err != nil {
		log.Fatalf("Could not create ID file: %v", err)
	}

	// Step 1. Initialize Pipeline structure
	p := pipeline.NewPipeline(threads, geneDir)

	// Step 2. Read Tree files
	files, err := ioutil.ReadDir(treeFilePath)
	if err != nil {
		log.Fatalf("Could not read tree dir %s: %v", treeFilePath, err)
	}

	var originalTreeFile string
	var treeContent string
	if len(files) == 1 {
		originalTreeFile = filepath.Join(treeFilePath, files[0].Name())
		fileBytes, err := ioutil.ReadFile(originalTreeFile)
		if err != nil {
			log.Fatalf("Could not read tree file: %v", err)
		}
		treeContent = string(fileBytes)
	} else {
		log.Fatalf("Expected exactly 1 tree file in %s, found %d", treeFilePath, len(files))
	}

	// Copy to temp directory
	actualTreeFile := filepath.Join(p.TempDir, filepath.Base(originalTreeFile))
	ioutil.WriteFile(actualTreeFile, []byte(treeContent), 0644)

	pattern1 := regexp.MustCompile(`\(([^()\s,]+),([^()\s,]+)\),\(([^()\s,]+),([^()\s,]+)\)`)
	pattern2 := regexp.MustCompile(`\(([^()\s,]+),([^()\s,]+)\),([^()\s,]+)`)
	pattern3 := regexp.MustCompile(`([^()\s,]+),\(([^()\s,]+),([^()\s,]+)\)`)

	nodeCounter := 1

	for {
		matches1 := pattern1.FindAllStringSubmatch(treeContent, -1)
		matches2 := pattern2.FindAllStringSubmatch(treeContent, -1)
		matches3 := pattern3.FindAllStringSubmatch(treeContent, -1)

		if len(matches1) == 0 && len(matches2) == 0 && len(matches3) == 0 {
			log.Println("Tree reduction complete. No more resolvable groups.")
			break
		}

		type taskDef struct {
			kind string
			m    []string
			id   string
		}

		var tasks []taskDef
		for _, m := range matches1 {
			tasks = append(tasks, taskDef{kind: "1", m: m, id: fmt.Sprintf("AG_%04d", nodeCounter)})
			nodeCounter++
		}
		for _, m := range matches2 {
			tasks = append(tasks, taskDef{kind: "2", m: m, id: fmt.Sprintf("AG_%04d", nodeCounter)})
			nodeCounter++
		}
		for _, m := range matches3 {
			tasks = append(tasks, taskDef{kind: "3", m: m, id: fmt.Sprintf("AG_%04d", nodeCounter)})
			nodeCounter++
		}

		log.Printf("Found %d resolvable task groups in this iteration", len(tasks))

		var wg sync.WaitGroup
		sem := make(chan struct{}, p.NumThreads) // concurrency limit

		var mu sync.Mutex
		replacements := make(map[string]string)
		var resultsLogs []string

		for _, t := range tasks {
			wg.Add(1)
			go func(task taskDef) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				var err error
				var strre string
				var resultLog string

				switch task.kind {
				case "1":
					log.Printf("Running BactAG2 for group %s + %s Outside %s + %s -> %s", task.m[1], task.m[2], task.m[3], task.m[4], task.id)
					err = p.RunBactAG2(task.m[1], task.m[2], task.m[3], task.m[4], task.id)
					strre = "(" + task.m[1] + "," + task.m[2] + ")"
					resultLog = fmt.Sprintf("%s+%s OutsidE %s+%s = %s", task.m[1], task.m[2], task.m[3], task.m[4], task.id)
				case "2":
					log.Printf("Running BactAG1 for group %s + %s Outside %s -> %s", task.m[1], task.m[2], task.m[3], task.id)
					err = p.RunBactAG1(task.m[1], task.m[2], task.m[3], task.id)
					strre = "(" + task.m[1] + "," + task.m[2] + ")"
					resultLog = fmt.Sprintf("%s+%s OutsidE %s = %s", task.m[1], task.m[2], task.m[3], task.id)
				case "3":
					log.Printf("Running BactAG1 for group %s + %s Outside %s -> %s", task.m[2], task.m[3], task.m[1], task.id)
					err = p.RunBactAG1(task.m[2], task.m[3], task.m[1], task.id)
					strre = "(" + task.m[2] + "," + task.m[3] + ")"
					resultLog = fmt.Sprintf("%s+%s OutsidE %s = %s", task.m[2], task.m[3], task.m[1], task.id)
				}

				if err != nil {
					log.Printf("Error processing %s: %v", task.id, err)
				}

				mu.Lock()
				replacements[strre] = task.id
				resultsLogs = append(resultsLogs, resultLog)
				mu.Unlock()

			}(t)
		}

		wg.Wait()

		// Write results to bactID.txt
		f, err := os.OpenFile(idFilePath, os.O_APPEND|os.O_WRONLY, 0644)
		if err == nil {
			for _, l := range resultsLogs {
				f.WriteString(l + "\n")
			}
			f.Close()
		}

		// Replace strings in tree content
		for strre, id := range replacements {
			treeContent = strings.Replace(treeContent, strre, id, -1)
		}

		// Overwrite tree file
		ioutil.WriteFile(actualTreeFile, []byte(treeContent), 0644)
		
		log.Printf("Iteration completed. Updated tree structure.")
	}

	log.Printf("All pipeline tasks completed.")

	// Post-Execution Lifecycle (from BACT_AG.sh)
	agOutput := fmt.Sprintf("BactAG-output%s", currTime)
	log.Printf("Packaging results into %s ...", agOutput)
	
	if err := os.MkdirAll(agOutput, 0755); err != nil {
		log.Printf("Failed to create output bundle dir: %v", err)
	}

	// Copy result folders (output, log, bactID.txt, tree_file)
	// Using exec "cp" to maintain simplicity and deep recursive fidelity for folders
	exec.Command("cp", "-r", "output", agOutput).Run()
	exec.Command("cp", "-r", "log", agOutput).Run()
	exec.Command("cp", filepath.Dir(treeFilePath), agOutput).Run()
	exec.Command("cp", idFilePath, agOutput).Run()

	// Clean up staging
	os.RemoveAll("output")
	os.RemoveAll("log")
	os.Remove(idFilePath)

	log.Printf("All done! All Output files and logs are in %s", agOutput)
}
