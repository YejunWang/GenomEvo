package cmd

import (
"bufio"
"fmt"
"io"
"os"
"os/exec"
"path/filepath"
"strings"

"github.com/spf13/cobra"
)

var (
refStrain string
cdHitC    string
cdHitS    string
bactCG1   string
bactCG2   string
inputDirRun  string
outputDirRun string
)

var runCmd = &cobra.Command{
Use:   "run",
Short: "Run the full pipeline with CD-HIT, BLAST, and optionally QC filtering",
Long:  `Executes the main BactCG pipeline. Integrates QC, CD-HIT redundancy clustering, and BactCG computation.`,
Run: func(cmd *cobra.Command, args []string) {
if !checkDependencies() {
os.Exit(1)
}

runQC := askForQC()
if runQC {
fmt.Println("[QC] Running QC: Filtering protein files smaller than 90% of the average size...")
if err := FilterFiles(inputDirRun); err != nil {
fmt.Printf("Error during QC filter: %v\n", err)
} else {
fmt.Println("[QC] QC Completed successfully.")
}
} else {
fmt.Println("[QC] Skipping QC filtering step.")
}

fmt.Println("Proceeding with CD-HIT clustering...")
out1 := filepath.Join(outputDirRun, "1.cd-hit_output")
out2 := filepath.Join(out1, "1.cd-hit_fatsa") // Spelled as in original script
os.MkdirAll(out1, 0755)
os.MkdirAll(out2, 0755)

files, err := os.ReadDir(inputDirRun)
if err != nil {
fmt.Printf("Error reading input directory: %v\n", err)
os.Exit(1)
}

		cdHitPath := "Tool/cd-hit-v4.6.7-2017-0501/cd-hit"
		if path, err := exec.LookPath("cd-hit"); err == nil {
			cdHitPath = path
		}
for _, f := range files {
if f.IsDir() {
continue
}
name := f.Name()
if !strings.HasSuffix(name, ".fasta") && !strings.HasSuffix(name, ".faa") && !strings.HasSuffix(name, ".fa") {
continue
}
baseName := strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(name, ".fasta"), ".faa"), ".fa")
inFile := filepath.Join(inputDirRun, name)
outFile := filepath.Join(out1, baseName+"_cd-hit.fasta")

fmt.Printf("Running CD-HIT for %s...\n", f.Name())
cdCmd := exec.Command(cdHitPath, "-i", inFile, "-o", outFile, "-c", cdHitC, "-M", "0", "-T", "0", "-s", cdHitS, "-d", "1000000")
if _, err := cdCmd.CombinedOutput(); err != nil {
fmt.Printf("Error running CD-HIT on %s: %v\n", f.Name(), err)
}

// Copy file to the final directory
newFaaFileName := filepath.Join(out2, baseName+".fasta")
copyFile(outFile, newFaaFileName)
}

fmt.Println("Proceeding with BactCG execution...")
// Running the cg sub-command natively via os.Args[0] (the executable itself)
exePath, _ := os.Executable()
cgCmd := exec.Command(exePath, "cg", out2, refStrain, bactCG1, bactCG2)
cgCmd.Stdout = os.Stdout
cgCmd.Stderr = os.Stderr
if err := cgCmd.Run(); err != nil {
fmt.Printf("Error running BactCG: %v\n", err)
}

// Copy result and clean up
cgRes := filepath.Join("result", "cg_result", "CG.tab.txt")
finalOut := filepath.Join(outputDirRun, "CG_ALL.txt")
if err := copyFile(cgRes, finalOut); err != nil {
fmt.Printf("Failed to copy final CG result: %v\n", err)
} else {
fmt.Printf("Successfully generated final result: %s\n", finalOut)
}

fmt.Println("Cleaning up temporary result folder...")
os.RemoveAll("result")
fmt.Println("Pipeline execution finished.")
},
}

func init() {
rootCmd.AddCommand(runCmd)
runCmd.Flags().StringVarP(&refStrain, "ref", "r", "NRRL_3357", "Reference strain name for BactCG")
runCmd.Flags().StringVar(&cdHitC, "cd-c", "0.7", "CD-HIT sequence identity threshold (-c)")
runCmd.Flags().StringVar(&cdHitS, "cd-s", "0.7", "CD-HIT length difference cutoff (-s)")
runCmd.Flags().StringVar(&bactCG1, "cg1", "0.8", "BactCG first parameter")
runCmd.Flags().StringVar(&bactCG2, "cg2", "0.9", "BactCG second parameter")
runCmd.Flags().StringVarP(&inputDirRun, "input", "i", "input/seq/prot_file/", "Input directory of protein files")
runCmd.Flags().StringVarP(&outputDirRun, "output", "o", "output/CG_results/", "Output directory for the results")
}

func checkDependencies() bool {
_, err := exec.LookPath("blastn")
if err != nil {
_, err2 := exec.LookPath("blastp")
if err2 != nil {
fmt.Println("Error: BLAST suite (blastn/blastp) is not found in your system PATH. Please install NCBI BLAST+ first!")
return false
}
}

cdHitPath := "Tool/cd-hit-v4.6.7-2017-0501/cd-hit"
	if path, err := exec.LookPath("cd-hit"); err == nil {
		cdHitPath = path
	}

	if _, err := os.Stat(cdHitPath); os.IsNotExist(err) {
		fmt.Printf("Error: CD-HIT executable not found at expected path: %s or in PATH.\nPlease ensure CD-HIT is installed and configured identically to BLAST as it is required to run.\n", cdHitPath)
		return false
	} else {
		fmt.Println("Dependency Check: CD-HIT found at:", cdHitPath)
	}

	clustalw2Path := "Tool/clustalw2"
	if path, err := exec.LookPath("clustalw2"); err == nil {
		clustalw2Path = path
	}

	if _, err := os.Stat(clustalw2Path); os.IsNotExist(err) {
		fmt.Printf("Error: Clustalw2 executable not found in PATH or at %s.\nPlease install clustalw2 to proceed.\n", clustalw2Path)
		return false
}

fmt.Println("Dependency Check: BLAST, CD-HIT, and Clustalw2 dependencies satisfied.")
return true
}

func askForQC() bool {
reader := bufio.NewReader(os.Stdin)
for {
fmt.Print("Do you want to run the QC (Quality Control filter for prot_files based on size)? [yes/no]: ")
input, err := reader.ReadString('\n')
if err != nil {
return false
}
input = strings.TrimSpace(strings.ToLower(input))
if input == "yes" || input == "y" {
return true
} else if input == "no" || input == "n" {
return false
}
fmt.Println("Please type 'yes' or 'no'.")
}
}

func copyFile(src, dst string) error {
in, err := os.Open(src)
if err != nil {
return err
}
defer in.Close()

out, err := os.Create(dst)
if err != nil {
return err
}
defer out.Close()

_, err = io.Copy(out, in)
return err
}
