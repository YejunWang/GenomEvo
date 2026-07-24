package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func runPipeline(args []string) {
	pipeCmd := flag.NewFlagSet("pipeline", flag.ExitOnError)
	gbkFile := pipeCmd.String("gbk", "", "Path to the input GenBank file (.gbk) (required)")
	pgFile := pipeCmd.String("pg", "", "Path to the Pan-Genome reference file (e.g. 26_PG.txt) (required)")
	seqDir := pipeCmd.String("seq", "", "Directory containing reference protein .faa sequences (required)")
	cgBin := pipeCmd.String("cg", "./bactcg", "Path to the bactcg (BactCG 2.0) executable")
	mode := pipeCmd.String("mode", "PGAG", "Annotation mode: PGAG or RAST")
	cov1 := pipeCmd.String("cov1", "0.7", "Coverage 1 (similarity) for CG")
	cov2 := pipeCmd.String("cov2", "0.7", "Coverage 2 (coverage) for CG")
	outDir := pipeCmd.String("out", "./output", "Output directory")

	if err := pipeCmd.Parse(args); err != nil {
		fmt.Println("Usage: bactpga pipeline -gbk <in.gbk> -pg <ref_PG.txt> -seq <ref_seq_dir> [options]")
		os.Exit(1)
	}

	if *gbkFile == "" || *pgFile == "" || *seqDir == "" {
		fmt.Println("Missing required flags for pipeline.")
		pipeCmd.Usage()
		os.Exit(1)
	}

	// 1. Prepare environment
	strainName := strings.TrimSuffix(filepath.Base(*gbkFile), filepath.Ext(*gbkFile))
	os.MkdirAll(*outDir, 0755)

	fmt.Printf("[1/3] Parsing GenBank file: %s for strain %s...\n", *gbkFile, strainName)
	tabOutPath := filepath.Join(*outDir, strainName+"_PGAG.tab.txt")
	err := runParseToFile(*gbkFile, tabOutPath)
	if err != nil {
		fmt.Printf("Parsing failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("[2/3] Running CG alignment using %s...\n", *cgBin)
	// Resolve CG binary to absolute path before changing working directory
	absCgBin, err := filepath.Abs(*cgBin)
	if err != nil {
		absCgBin = *cgBin
	}

	// Prepare a dedicated working directory for CG execution
	cgWorkDir := filepath.Join(*outDir, "cg_work")
	cgSeqDir := filepath.Join(cgWorkDir, "seq")
	os.RemoveAll(cgWorkDir)
	os.MkdirAll(cgSeqDir, 0755)

	// Copy protein sequences (.faa files) to CG work dir under seq/
	// The new bactcg expects .fasta extension; copy .faa files to .fasta
	entries, _ := os.ReadDir(*seqDir)
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".faa") {
			src := filepath.Join(*seqDir, entry.Name())
			dstName := strings.TrimSuffix(entry.Name(), ".faa") + ".fasta"
			dst := filepath.Join(cgSeqDir, dstName)
			data, err := os.ReadFile(src)
			if err == nil {
				os.WriteFile(dst, data, 0644)
			}
		} else if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".fasta") {
			src := filepath.Join(*seqDir, entry.Name())
			dst := filepath.Join(cgSeqDir, entry.Name())
			data, err := os.ReadFile(src)
			if err == nil {
				os.WriteFile(dst, data, 0644)
			}
		}
	}

	// Execute new BactCG: bactcg cg <seq_dir> <ref> <similarity> <coverage>
	cgCmd := exec.Command(absCgBin, "cg", "seq", strainName, *cov1, *cov2)
	cgCmd.Dir = cgWorkDir
	cgCmd.Stdout = os.Stdout
	cgCmd.Stderr = os.Stderr
	if err := cgCmd.Run(); err != nil {
		fmt.Printf("CG alignment failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("[3/3] Annotating parsed table with PG UIDs using MBH...\n")
	mutbestDir := filepath.Join(cgWorkDir, "result", "out_mutbest_filt")
	finalOutPath := filepath.Join(*outDir, strainName+"_PGAG_PGA.tab.txt")
	
	err = runAnnotateToFile(*pgFile, tabOutPath, *mode, strainName, mutbestDir, "", finalOutPath)
	if err != nil {
		fmt.Printf("Annotation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Pipeline completed successfully! Final output saved to: %s\n", finalOutPath)
}

func runParseToFile(gbFileName, outFileName string) error {
	saveStdout := os.Stdout
	out, err := os.Create(outFileName)
	if err != nil {
		return err
	}
	defer out.Close()
	os.Stdout = out
	runParse(gbFileName)
	os.Stdout = saveStdout
	return nil
}

func runAnnotateToFile(pgmapmatFileName, agFileName, mode, strain, mutbestDir, seqDir, outFileName string) error {
	saveStdout := os.Stdout
	out, err := os.Create(outFileName)
	if err != nil {
		return err
	}
	defer out.Close()
	os.Stdout = out
	runAnnotate(pgmapmatFileName, agFileName, mode, strain, mutbestDir, seqDir)
	os.Stdout = saveStdout
	return nil
}
