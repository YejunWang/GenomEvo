package cmd

import (
"bufio"
"fmt"
"os"
"path/filepath"
"strings"
"github.com/spf13/cobra"
)

var (
inputCdhitRun   string
inputCgFileRun  string
baseOutDirRun   string
)

var getFaCmd = &cobra.Command{
Use:   "getfa",
Short: "Extract sequences from CD-HIT fasta to independent .fa files per family",
Run: func(cmd *cobra.Command, args []string) {
err := GetFaFileAll(inputCdhitRun, inputCgFileRun, baseOutDirRun)
if err != nil {
fmt.Println("Error:", err)
}
},
}

func init() {
rootCmd.AddCommand(getFaCmd)
getFaCmd.Flags().StringVarP(&inputCdhitRun, "cdhit", "c", "output/CG_results/1.cd-hit_output/1.cd-hit_fatsa", "Directory with cdhit FASTA")
getFaCmd.Flags().StringVarP(&inputCgFileRun, "cgfile", "g", "output/CG_results/CG_ALL.txt", "CG core gene table file")
getFaCmd.Flags().StringVarP(&baseOutDirRun, "outdir", "o", "output/CG_results/all-strain-together", "Base output directory")
}

func GetFaFileAll(inputCdhitDir, cgFile, baseOutDir string) error {
fmt.Println("Running Post-process Step 2: get_fa_file-all...")
outDir := filepath.Join(baseOutDir, "2.result")
os.MkdirAll(outDir, 0755)

seqDict := make(map[string]string)
files, err := os.ReadDir(inputCdhitDir)
if err != nil {
return fmt.Errorf("failed to read cdhit dir: %v", err)
}

for _, f := range files {
if !f.IsDir() && strings.HasSuffix(f.Name(), ".fasta") {
indexName := strings.TrimSuffix(f.Name(), ".fasta")
path := filepath.Join(inputCdhitDir, f.Name())
file, err := os.Open(path)
if err != nil {
continue
}
scanner := bufio.NewScanner(file)
var currentSeqName string
for scanner.Scan() {
line := scanner.Text()
if strings.HasPrefix(line, ">") {
currentSeqName = ">" + indexName + "---" + strings.TrimPrefix(line, ">")
seqDict[currentSeqName] = ""
} else if line != "" {
seqDict[currentSeqName] += strings.TrimSpace(line)
}
}
file.Close()
}
}

cgLineOutput := make([][]string, 0)
cgf, err := os.Open(cgFile)
if err != nil {
return fmt.Errorf("failed to open CG file: %v", err)
}
defer cgf.Close()

scanner := bufio.NewScanner(cgf)
if !scanner.Scan() {
return nil
}
firstLine := scanner.Text()
headers := strings.Split(firstLine, "\t")
var indexList []string
for _, h := range headers {
indexList = append(indexList, strings.TrimSpace(h)+"---")
}

for scanner.Scan() {
line := scanner.Text()
if line == "" {
continue
}
parts := strings.Split(line, "\t")
var outLine []string
for i := 0; i < len(parts) && i < len(indexList); i++ {
outLine = append(outLine, indexList[i]+strings.TrimSpace(parts[i]))
}
cgLineOutput = append(cgLineOutput, outLine)
}

for _, row := range cgLineOutput {
if len(row) == 0 {
continue
}
outFileName := strings.TrimPrefix(row[0], indexList[0])
outPath := filepath.Join(outDir, outFileName+".fa")
outf, err := os.Create(outPath)
if err != nil {
continue
}
writer := bufio.NewWriter(outf)
for _, el := range row {
key := ">" + el
if val, ok := seqDict[key]; ok {
writer.WriteString(key + "\n" + val + "\n")
}
}
writer.Flush()
outf.Close()
}
return nil
}

func GetSnpMegaAll(inputMegaDir, outputDir string) error {
	fmt.Println("Running Post-process Step 4: get_SNP_mega-all...")
	os.MkdirAll(outputDir, 0755)

	// Step 4a: For each .meg file, remove identical (non-informative) sites
	newMegaDir := filepath.Join(outputDir, "new_mega")
	os.MkdirAll(newMegaDir, 0755)

	files, err := os.ReadDir(inputMegaDir)
	if err != nil {
		return fmt.Errorf("failed to read mega dir: %v", err)
	}

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".meg") {
			continue
		}

		inPath := filepath.Join(inputMegaDir, f.Name())
		outPath := filepath.Join(newMegaDir, "new-"+f.Name())

		if err := filterIdenticalSites(inPath, outPath); err != nil {
			fmt.Printf("Warning: SNP filter failed for %s: %v\n", f.Name(), err)
		}
	}

	// Step 4b: Concatenate all filtered MEGA files
	return concatMegaFiles(newMegaDir, outputDir)
}

// filterIdenticalSites removes positions where all sequences have identical amino acids.
func filterIdenticalSites(inFile, outFile string) error {
	in, err := os.Open(inFile)
	if err != nil {
		return err
	}
	defer in.Close()

	// Read sequences
	seqs := make(map[string]string)
	var seqOrder []string
	scanner := bufio.NewScanner(in)
	var currentId string
	inHeader := true

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			if inHeader && !strings.HasPrefix(line, "#MEGA") && !strings.HasPrefix(line, "!") {
				// Sequence ID line
				currentId = strings.TrimPrefix(line, "#")
				seqOrder = append(seqOrder, currentId)
				seqs[currentId] = ""
				inHeader = false
			}
			continue
		}
		if strings.HasPrefix(line, ">") {
			currentId = strings.TrimPrefix(line, ">")
			seqOrder = append(seqOrder, currentId)
			seqs[currentId] = ""
			continue
		}
		if currentId != "" {
			seqs[currentId] += line
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if len(seqOrder) == 0 || len(seqs[seqOrder[0]]) == 0 {
		// No sequences or empty - copy as-is
		return copyFile(inFile, outFile)
	}

	// Find positions where all sequences have identical AA
	seqLen := len(seqs[seqOrder[0]])
	keepPos := make([]bool, seqLen)
	for i := 0; i < seqLen; i++ {
		firstAA := seqs[seqOrder[0]][i]
		allSame := true
		for _, id := range seqOrder[1:] {
			if i >= len(seqs[id]) || seqs[id][i] != firstAA {
				allSame = false
				break
			}
		}
		keepPos[i] = !allSame // keep if NOT all identical
	}

	// Write filtered output
	out, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer out.Close()

	writer := bufio.NewWriter(out)
	writer.WriteString("#MEGA\n!Title CG;\n!Format DataType=Protein indel=-;\n\n")

	for _, id := range seqOrder {
		writer.WriteString(fmt.Sprintf("#%s\n", id))
		filtered := ""
		for i := 0; i < seqLen && i < len(seqs[id]); i++ {
			if keepPos[i] {
				filtered += string(seqs[id][i])
			}
		}
		writer.WriteString(filtered + "\n\n")
	}
	writer.Flush()
	return nil
}

// concatMegaFiles concatenates all filtered MEGA files into a single SNP alignment.
func concatMegaFiles(inputDir, outputDir string) error {
	files, err := os.ReadDir(inputDir)
	if err != nil {
		return fmt.Errorf("failed to read new_mega dir: %v", err)
	}

	allSeqs := make(map[string]string)
	var indexList []string
	var initLen int

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".meg") {
			continue
		}

		path := filepath.Join(inputDir, f.Name())
		file, err := os.Open(path)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(file)
		isFirst := (len(indexList) == 0)
		var currentId string

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			if strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "#MEGA") && !strings.HasPrefix(line, "!") {
				currentId = strings.TrimPrefix(line, "#")
				if isFirst {
					indexList = append(indexList, currentId)
					allSeqs[currentId] = ""
				}
			} else if !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "!") {
				if _, ok := allSeqs[currentId]; ok {
					allSeqs[currentId] += line
				}
			}
		}
		file.Close()

		if isFirst && len(indexList) > 0 {
			initLen = len(allSeqs[indexList[0]])
		}
	}

	outPath := filepath.Join(outputDir, "all_core_gene.meg")
	outf, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outf.Close()

	writer := bufio.NewWriter(outf)
	writer.WriteString("#MEGA\n!Title SNPMega;\n!Format DataType=Protein indel=-;\n\n")

	for _, id := range indexList {
		writer.WriteString(fmt.Sprintf("#%s\n", id))
		writer.WriteString(allSeqs[id] + "\n\n")
	}
	writer.Flush()

	fmt.Printf("Step 4 output to %s with concatenated length %d\n", outPath, initLen)
	return nil
}
// --- snpmega subcommand ---

var (
	snpmegaInputDir  string
	snpmegaOutputDir string
)

var snpmegaCmd = &cobra.Command{
	Use:   "snpmega",
	Short: "Concatenate MEGA alignments into SNP matrix",
	Long:  `Port of get_SNP_mega-all.py. Reads all .meg files in input dir, removes identical sites, concatenates sequences, and outputs a single all_core_gene.meg file.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := GetSnpMegaAll(snpmegaInputDir, snpmegaOutputDir); err != nil {
			fmt.Println("Error running snpmega:", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(snpmegaCmd)
	snpmegaCmd.Flags().StringVarP(&snpmegaInputDir, "input", "i", "output/CG_results/all-strain-together/3.mega", "Directory containing .meg files")
	snpmegaCmd.Flags().StringVarP(&snpmegaOutputDir, "output", "o", "output/CG_results/all-strain-together/4.SNP_mega", "Output directory for SNP mega file")
}