package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "parse":
		parseCmd := flag.NewFlagSet("parse", flag.ExitOnError)
		if err := parseCmd.Parse(os.Args[2:]); err != nil || parseCmd.NArg() < 1 {
			fmt.Println("Usage: bactpga parse <genbank.gb>")
			os.Exit(1)
		}
		gbFileName := parseCmd.Arg(0)
		runParse(gbFileName)
	case "annotate":
		annotateCmd := flag.NewFlagSet("annotate", flag.ExitOnError)
		pgFile := annotateCmd.String("pg", "", "Path to the 26_PG.txt file (required)")
		agFile := annotateCmd.String("tab", "", "Path to the annotation table file (required)")
		mode := annotateCmd.String("mode", "", "Annotation mode: PGAG or RAST (required)")
		strain := annotateCmd.String("strain", "", "Target strain name (required)")
		mutbestDir := annotateCmd.String("mutbestDir", ".", "Directory containing .mutbest.filt.txt files")
		seqDir := annotateCmd.String("seqDir", "", "(Optional) Directory containing .faa sequences to count total alignments. Default will guess by files in mutbestDir.")
		
		if err := annotateCmd.Parse(os.Args[2:]); err != nil {
			fmt.Println("Usage: bactpga annotate -pg <pg.txt> -tab <ag.tab> -mode <PGAG|RAST> -strain <strain> [options]")
			os.Exit(1)
		}
		if *pgFile == "" || *agFile == "" || *mode == "" || *strain == "" {
			fmt.Println("Missing required flags for annotate.")
			annotateCmd.Usage()
			os.Exit(1)
		}
		runAnnotate(*pgFile, *agFile, *mode, *strain, *mutbestDir, *seqDir)
	case "pipeline":
		runPipeline(os.Args[2:])
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("BactPGA - A combined tool for parsing GenBank and annotating Salmonella PG UIDs")
	fmt.Println("\nUsage:")
	fmt.Println("  bactpga <command> [arguments]")
	fmt.Println("\nCommands:")
	fmt.Println("  parse     Extract features (CDS, rRNA, etc.) from GenBank files (.gbk/.gb)")
	fmt.Println("  annotate  Annotate gene tables with Salmonella PG UID using mutual best hits")
	fmt.Println("  pipeline  Run the complete automated pipeline (Parse -> CG -> Annotate)")
	fmt.Println("\nRun 'bactpga <command> -h' for more details.")
}
