package pipeline

import (
	"bactag/internal/tools"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Pipeline struct holds configuration for the BactAG run
type Pipeline struct {
	NumThreads int
	GeneDir    string
	TempDir    string
	OutputDir  string
	LogDir     string
}

// NewPipeline initializes a new Pipeline workspace
func NewPipeline(threads int, geneDir string) *Pipeline {
	p := &Pipeline{
		NumThreads: threads,
		GeneDir:    geneDir,
		TempDir:    "temp",
		OutputDir:  "output",
		LogDir:     "log",
	}
	os.MkdirAll(p.TempDir, 0755)
	os.MkdirAll(p.OutputDir, 0755)
	os.MkdirAll(p.LogDir, 0755)
	return p
}

// runPipelineCmd parses a bash-like command to support I/O redirection natively without bash.
func (p *Pipeline) runPipelineCmd(dir, cmdStr string, logger io.Writer) error {
	// Simple tokenize that respects spaces but handles > 
	tokens := strings.Fields(cmdStr)
	if len(tokens) == 0 {
		return nil
	}

	var args []string
	var outFile string

	for i := 0; i < len(tokens); i++ {
		if strings.HasPrefix(tokens[i], ">") {
			if len(tokens[i]) > 1 {
				outFile = tokens[i][1:]
			} else if i+1 < len(tokens) {
				outFile = tokens[i+1]
				i++
			}
		} else if tokens[i] == ">" && i+1 < len(tokens) {
			outFile = tokens[i+1]
			i++
		} else {
			args = append(args, tokens[i])
		}
	}

	name := args[0]
	runArgs := args[1:]
	
	internalTools := map[string]bool{
		"homBB":            true,
		"homBlkReorder":    true,
		"orthBB":           true,
		"orthJoin1":        true,
		"orthJoin2":        true,
		"orthJoin3":        true,
		"orthoCombine":     true,
		"orthoParsing":     true,
		"patching":         true,
		"patching1":        true,
		"progBackbonePrep": true,
		"revcomp":          true,
	}

	execCmd := name
	if internalTools[name] {
		ex, _ := os.Executable()
		execCmd = ex
		runArgs = append([]string{name}, args[1:]...)
	}

	cmd := exec.Command(execCmd, runArgs...)
	cmd.Dir = dir

	if outFile != "" {
		f, err := os.Create(filepath.Join(dir, outFile))
		if err != nil {
			return err
		}
		defer f.Close()
		cmd.Stdout = f
	} else {
		cmd.Stdout = os.Stdout
	}
	cmd.Stderr = os.Stderr

	fmt.Printf("[EXEC] %s\n", cmdStr)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command failed: %s %v", cmdStr, err)
	}
	return nil
}

// RunBactAG1 executes the logic for BactAG1 integration
func (p *Pipeline) RunBactAG1(input1, input2, input3, output string) error {
	processDir := filepath.Join(p.TempDir, output)
	os.MkdirAll(processDir, 0755)

	// Step 1: Resolve and copy fasta files. 
	// If a fasta is not found in input/gene, it checks the output/ directory (derived from previous nodes).
	for _, inp := range []string{input1, input2, input3} {
		fname := inp + ".fasta"
		srcPath := filepath.Join(p.GeneDir, fname)
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			srcPath = filepath.Join(p.OutputDir, fname)
		}
		dstPath := filepath.Join(processDir, fname)
		if err := CopyFile(srcPath, dstPath); err != nil {
			return fmt.Errorf("failed to copy dependency %s: %v", srcPath, err)
		}
	}
	
	// Set log file
	logFile, err := os.Create(filepath.Join(p.LogDir, output+".log"))
	if err != nil {
		return err
	}
	defer logFile.Close()

	// Redirect output locally (simplified for demonstration)
	fmt.Fprintf(logFile, "Starting BactAG1 for %s, %s, %s -> %s\n", input1, input2, input3, output)

	// Example: (1) AG backbone inference and refinement
	fmt.Println("(1) AG backbone inference and refinement ...")
	
	// External commands
	p.runPipelineCmd(processDir, fmt.Sprintf("progressiveMauve --output=%s.vs.%s.xmfa --output-guide-tree=%s.vs.%s.guide_tree --backbone-output=%s.vs.%s.backbone %s.fasta %s.fasta", input1, input2, input1, input2, input1, input2, input1, input2, logFile), logFile)
	
	p.runPipelineCmd(processDir, fmt.Sprintf("progBackbonePrep %s.vs.%s.backbone >%s.vs.%s.backbone.txt", input1, input2, input1, input2, logFile), logFile)
	p.runPipelineCmd(processDir, fmt.Sprintf("homBlkReorder %s.vs.%s.backbone.txt >%s.vs.%s.homblk.txt", input1, input2, input1, input2, logFile), logFile)
	p.runPipelineCmd(processDir, fmt.Sprintf("orthoParsing %s.vs.%s.homblk.txt >%s.vs.%s.orthBlk.txt", input1, input2, input1, input2, logFile), logFile)
	
	p.runPipelineCmd(processDir, fmt.Sprintf("orthoCombine %s.vs.%s.orthBlk.txt 1000 >%s.vs.%s.combined_orthBlk.txt", input1, input2, input1, input2, logFile), logFile)
	p.runPipelineCmd(processDir, fmt.Sprintf("orthJoin1 %s.vs.%s.combined_orthBlk.txt %s.vs.%s.combined_orthBlk.txt %s %s >%s.vs.%s.Joint.txt", input1, input2, input1, input2, input1, input1, input1, input2, logFile), logFile)
	
	p.runPipelineCmd(processDir, fmt.Sprintf("homBB %s.vs.%s.Joint.txt %s >Houtenae_BG0_HOM%s.txt", input1, input2, input1, output, logFile), logFile)
	p.runPipelineCmd(processDir, fmt.Sprintf("orthBB Houtenae_BG0_HOM%s.txt >Houtenae_BG0_ORTH%s.txt", output, output, logFile), logFile)
	
	// Internal Tool Call Instead of perl script!
	orthoFile := filepath.Join(processDir, fmt.Sprintf("Houtenae_BG0_ORTH%s.txt", output))
	outFasta := filepath.Join(processDir, fmt.Sprintf("Houtenae_BG0_ORTH%s.fasta", output))
	err = tools.ExtractSequences(input1, filepath.Join(processDir, input1+".fasta"), input2, filepath.Join(processDir, input2+".fasta"), orthoFile, outFasta)
	if err != nil {
		fmt.Printf("ExtractSequences failed: %v\n", err)
	}

	// (2) Round 2 with Patching...
	fmt.Println("(2) Recursively patching AG...")
	// External commands for Round 2 patching
	p.runPipelineCmd(processDir, fmt.Sprintf("progressiveMauve --output=%s.vs.Houtenae_BG0_ORTH%s.xmfa --output-guide-tree=%s.vs.Houtenae_BG0_ORTH%s.guide_tree --backbone-output=%s.vs.Houtenae_BG0_ORTH%s.backbone %s.fasta Houtenae_BG0_ORTH%s.fasta", input1, output, input1, output, input1, output, input1, output, logFile), logFile)
	p.runPipelineCmd(processDir, fmt.Sprintf("progressiveMauve --output=%s.vs.%s.xmfa --output-guide-tree=%s.vs.%s.guide_tree --backbone-output=%s.vs.%s.backbone %s.fasta %s.fasta", input1, input3, input1, input3, input1, input3, input1, input3, logFile), logFile)
	
	p.runPipelineCmd(processDir, fmt.Sprintf("progBackbonePrep %s.vs.Houtenae_BG0_ORTH%s.backbone >%s.vs.Houtenae_BG0_ORTH%s.backbone.txt", input1, output, input1, output, logFile), logFile)
	p.runPipelineCmd(processDir, fmt.Sprintf("progBackbonePrep %s.vs.%s.backbone >%s.vs.%s.backbone.txt", input1, input3, input1, input3, logFile), logFile)
	
	p.runPipelineCmd(processDir, fmt.Sprintf("homBlkReorder %s.vs.Houtenae_BG0_ORTH%s.backbone.txt >%s.vs.Houtenae_BG0_ORTH%s.homblk.txt", input1, output, input1, output, logFile), logFile)
	p.runPipelineCmd(processDir, fmt.Sprintf("homBlkReorder %s.vs.%s.backbone.txt >%s.vs.%s.homblk.txt", input1, input3, input1, input3, logFile), logFile)
	
	p.runPipelineCmd(processDir, fmt.Sprintf("orthoParsing %s.vs.Houtenae_BG0_ORTH%s.homblk.txt >%s.vs.Houtenae_BG0_ORTH%s.orthBlk.txt", input1, output, input1, output, logFile), logFile)
	p.runPipelineCmd(processDir, fmt.Sprintf("orthoParsing %s.vs.%s.homblk.txt >%s.vs.%s.orthBlk.txt", input1, input3, input1, input3, logFile), logFile)
	
	p.runPipelineCmd(processDir, fmt.Sprintf("patching %s.vs.Houtenae_BG0_ORTH%s.orthBlk.txt %s.vs.%s.orthBlk.txt Houtenae_BG0_ORTH%s.txt %s %s %s_added >Houtenae_BG1_ORTH1%s.txt", input1, output, input1, input3, output, input1, input3, input3, output, logFile), logFile)

	fileA := filepath.Join(processDir, fmt.Sprintf("%s.vs.Houtenae_BG0_ORTH%s.orthBlk.txt", input1, output))
	fileB := filepath.Join(processDir, fmt.Sprintf("%s.vs.%s.orthBlk.txt", input1, input3))
	fileC := filepath.Join(processDir, fmt.Sprintf("Houtenae_BG1_ORTH1%s.txt", output))
	fileD := filepath.Join(processDir, fmt.Sprintf("Houtenae_BG1_ORTH2%s.txt", output))
	tools.InsertIntoFile(fileA, fileB, fileC, fileD)

	fileOut := filepath.Join(processDir, fmt.Sprintf("Houtenae_BG1_ORTH%s.txt", output))
	tools.ProcessTiaozhen(fileD, fileOut, input1, input3, input2)

	// Retrieve temp seq
	orthoFileIter1 := filepath.Join(processDir, fmt.Sprintf("Houtenae_BG1_ORTH%s.txt", output))
	outFastaIter1 := filepath.Join(processDir, fmt.Sprintf("Houtenae_BG1_ORTH%s.fasta", output))
	tools.ExtractSequences(input1, filepath.Join(processDir, input1+".fasta"), input2, filepath.Join(processDir, input2+".fasta"), orthoFileIter1, outFastaIter1)

	// Round 3 with Patching...
	p.runPipelineCmd(processDir, fmt.Sprintf("progressiveMauve --output=%s.vs.Houtenae_BG1_ORTH%s.xmfa --output-guide-tree=%s.vs.Houtenae_BG1_ORTH%s.guide_tree --backbone-output=%s.vs.Houtenae_BG1_ORTH%s.backbone %s.fasta Houtenae_BG1_ORTH%s.fasta", input2, output, input2, output, input2, output, input2, output, logFile), logFile)
	p.runPipelineCmd(processDir, fmt.Sprintf("progressiveMauve --output=%s.vs.%s.xmfa --output-guide-tree=%s.vs.%s.guide_tree --backbone-output=%s.vs.%s.backbone %s.fasta %s.fasta", input2, input3, input2, input3, input2, input3, input2, input3, logFile), logFile)
	
	p.runPipelineCmd(processDir, fmt.Sprintf("progBackbonePrep %s.vs.Houtenae_BG1_ORTH%s.backbone >%s.vs.Houtenae_BG1_ORTH%s.backbone.txt", input2, output, input2, output, logFile), logFile)
	p.runPipelineCmd(processDir, fmt.Sprintf("progBackbonePrep %s.vs.%s.backbone >%s.vs.%s.backbone.txt", input2, input3, input2, input3, logFile), logFile)
	
	p.runPipelineCmd(processDir, fmt.Sprintf("homBlkReorder %s.vs.Houtenae_BG1_ORTH%s.backbone.txt >%s.vs.Houtenae_BG1_ORTH%s.homblk.txt", input2, output, input2, output, logFile), logFile)
	p.runPipelineCmd(processDir, fmt.Sprintf("homBlkReorder %s.vs.%s.backbone.txt >%s.vs.%s.homblk.txt", input2, input3, input2, input3, logFile), logFile)
	
	p.runPipelineCmd(processDir, fmt.Sprintf("orthoParsing %s.vs.Houtenae_BG1_ORTH%s.homblk.txt >%s.vs.Houtenae_BG1_ORTH%s.orthBlk.txt", input2, output, input2, output, logFile), logFile)
	p.runPipelineCmd(processDir, fmt.Sprintf("orthoParsing %s.vs.%s.homblk.txt >%s.vs.%s.orthBlk.txt", input2, input3, input2, input3, logFile), logFile)
	
	p.runPipelineCmd(processDir, fmt.Sprintf("patching %s.vs.Houtenae_BG1_ORTH%s.orthBlk.txt %s.vs.%s.orthBlk.txt Houtenae_BG1_ORTH%s.txt %s %s %s_added >Houtenae_BG2_ORTH1%s.txt", input2, output, input2, input3, output, input2, input3, input3, output, logFile), logFile)

	fileA2 := filepath.Join(processDir, fmt.Sprintf("%s.vs.Houtenae_BG1_ORTH%s.orthBlk.txt", input2, output))
	fileB2 := filepath.Join(processDir, fmt.Sprintf("%s.vs.%s.orthBlk.txt", input2, input3))
	fileC2 := filepath.Join(processDir, fmt.Sprintf("Houtenae_BG2_ORTH1%s.txt", output))
	fileD2 := filepath.Join(processDir, fmt.Sprintf("Houtenae_BG2_ORTH2%s.txt", output))
	tools.InsertIntoFile(fileA2, fileB2, fileC2, fileD2)

	fileOut2 := filepath.Join(processDir, fmt.Sprintf("Houtenae_BG2_ORTH%s.txt", output))
	tools.ProcessTiaozhen(fileD2, fileOut2, input2, input3, input1)

	// Final Output Gen
	orthoFileIter2 := filepath.Join(processDir, fmt.Sprintf("Houtenae_BG2_ORTH%s.txt", output))
	finalGeneratedFasta := filepath.Join(processDir, fmt.Sprintf("Houtenae_BG2_ORTH%s.fasta", output))
	err2 := tools.ExtractSequences(input2, filepath.Join(processDir, input2+".fasta"), input1, filepath.Join(processDir, input1+".fasta"), orthoFileIter2, finalGeneratedFasta)
	if err2 != nil {
		fmt.Printf("ExtractSequences error: %v\n", err2)
	}

	CopyFile(finalGeneratedFasta, filepath.Join(p.OutputDir, output+".fasta"))

	return nil
}

// RunBactAG2 executes the extended logic for BactAG2
func (p *Pipeline) RunBactAG2(input1, input2, input3, input4, output string) error {
processDir := filepath.Join(p.TempDir, output)
os.MkdirAll(processDir, 0755)

for _, inp := range []string{input1, input2, input3, input4} {
fname := inp + ".fasta"
srcPath := filepath.Join(p.GeneDir, fname)
if _, err := os.Stat(srcPath); os.IsNotExist(err) {
srcPath = filepath.Join(p.OutputDir, fname)
}
dstPath := filepath.Join(processDir, fname)
if err := CopyFile(srcPath, dstPath); err != nil {
return fmt.Errorf("failed to copy dependency %s: %v", srcPath, err)
}
}

logFile, err := os.Create(filepath.Join(p.LogDir, output+".log"))
if err != nil {
return err
}
defer logFile.Close()

fmt.Fprintf(logFile, "Starting BactAG2 for %s, %s, %s, %s -> %s\n", input1, input2, input3, input4, output)

fmt.Println("(1) AG backbone inference and refinement ...")

p.runPipelineCmd(processDir, fmt.Sprintf("progressiveMauve --output=%s.vs.%s.xmfa --output-guide-tree=%s.vs.%s.guide_tree --backbone-output=%s.vs.%s.backbone %s.fasta %s.fasta", input1, input2, input1, input2, input1, input2, input1, input2, logFile), logFile)
p.runPipelineCmd(processDir, fmt.Sprintf("progBackbonePrep %s.vs.%s.backbone >%s.vs.%s.backbone.txt", input1, input2, input1, input2, logFile), logFile)
p.runPipelineCmd(processDir, fmt.Sprintf("homBlkReorder %s.vs.%s.backbone.txt >%s.vs.%s.homblk.txt", input1, input2, input1, input2, logFile), logFile)
p.runPipelineCmd(processDir, fmt.Sprintf("orthoParsing %s.vs.%s.homblk.txt >%s.vs.%s.orthBlk.txt", input1, input2, input1, input2, logFile), logFile)
p.runPipelineCmd(processDir, fmt.Sprintf("orthoCombine %s.vs.%s.orthBlk.txt 1000 >%s.vs.%s.combined_orthBlk.txt", input1, input2, input1, input2, logFile), logFile)
p.runPipelineCmd(processDir, fmt.Sprintf("orthJoin1 %s.vs.%s.combined_orthBlk.txt %s.vs.%s.combined_orthBlk.txt %s %s >%s.vs.%s.Joint.txt", input1, input2, input1, input2, input1, input1, input1, input2, logFile), logFile)
p.runPipelineCmd(processDir, fmt.Sprintf("homBB %s.vs.%s.Joint.txt %s >Houtenae_BG0_HOM%s.txt", input1, input2, input1, output, logFile), logFile)
p.runPipelineCmd(processDir, fmt.Sprintf("orthBB Houtenae_BG0_HOM%s.txt >Houtenae_BG0_ORTH%s.txt", output, output, logFile), logFile)

orthoFile := filepath.Join(processDir, fmt.Sprintf("Houtenae_BG0_ORTH%s.txt", output))
outFasta := filepath.Join(processDir, fmt.Sprintf("Houtenae_BG0_ORTH%s.fasta", output))
err = tools.ExtractSequences(input1, filepath.Join(processDir, input1+".fasta"), input2, filepath.Join(processDir, input2+".fasta"), orthoFile, outFasta)

fmt.Println("1-st round Alignment ...")
p.runPipelineCmd(processDir, fmt.Sprintf("progressiveMauve --output=%s.vs.Houtenae_BG0_ORTH%s.xmfa --output-guide-tree=%s.vs.Houtenae_BG0_ORTH%s.guide_tree --backbone-output=%s.vs.Houtenae_BG0_ORTH%s.backbone %s.fasta Houtenae_BG0_ORTH%s.fasta", input1, output, input1, output, input1, output, input1, output, logFile), logFile)
p.runPipelineCmd(processDir, fmt.Sprintf("progressiveMauve --output=%s.vs.%s.xmfa --output-guide-tree=%s.vs.%s.guide_tree --backbone-output=%s.vs.%s.backbone %s.fasta %s.fasta", input1, input3, input1, input3, input1, input3, input1, input3, logFile), logFile)

p.runPipelineCmd(processDir, fmt.Sprintf("progBackbonePrep %s.vs.Houtenae_BG0_ORTH%s.backbone >%s.vs.Houtenae_BG0_ORTH%s.backbone.txt", input1, output, input1, output, logFile), logFile)
p.runPipelineCmd(processDir, fmt.Sprintf("progBackbonePrep %s.vs.%s.backbone >%s.vs.%s.backbone.txt", input1, input3, input1, input3, logFile), logFile)

p.runPipelineCmd(processDir, fmt.Sprintf("homBlkReorder %s.vs.Houtenae_BG0_ORTH%s.backbone.txt >%s.vs.Houtenae_BG0_ORTH%s.homblk.txt", input1, output, input1, output, logFile), logFile)
p.runPipelineCmd(processDir, fmt.Sprintf("homBlkReorder %s.vs.%s.backbone.txt >%s.vs.%s.homblk.txt", input1, input3, input1, input3, logFile), logFile)

p.runPipelineCmd(processDir, fmt.Sprintf("orthoParsing %s.vs.Houtenae_BG0_ORTH%s.homblk.txt >%s.vs.Houtenae_BG0_ORTH%s.orthBlk.txt", input1, output, input1, output, logFile), logFile)
p.runPipelineCmd(processDir, fmt.Sprintf("orthoParsing %s.vs.%s.homblk.txt >%s.vs.%s.orthBlk.txt", input1, input3, input1, input3, logFile), logFile)

p.runPipelineCmd(processDir, fmt.Sprintf("patching %s.vs.Houtenae_BG0_ORTH%s.orthBlk.txt %s.vs.%s.orthBlk.txt Houtenae_BG0_ORTH%s.txt %s %s %s_added >Houtenae_BG1_ORTH1%s.txt", input1, output, input1, input3, output, input1, input3, input3, output, logFile), logFile)

fileA := filepath.Join(processDir, fmt.Sprintf("%s.vs.Houtenae_BG0_ORTH%s.orthBlk.txt", input1, output))
fileB := filepath.Join(processDir, fmt.Sprintf("%s.vs.%s.orthBlk.txt", input1, input3))
fileC := filepath.Join(processDir, fmt.Sprintf("Houtenae_BG1_ORTH1%s.txt", output))
fileD := filepath.Join(processDir, fmt.Sprintf("Houtenae_BG1_ORTH2%s.txt", output))
tools.InsertIntoFile(fileA, fileB, fileC, fileD)

fileOut := filepath.Join(processDir, fmt.Sprintf("Houtenae_BG1_ORTH%s.txt", output))
tools.ProcessTiaozhen(fileD, fileOut, input1, input3, input2)

orthoFileIter1 := filepath.Join(processDir, fmt.Sprintf("Houtenae_BG1_ORTH%s.txt", output))
outFastaIter1 := filepath.Join(processDir, fmt.Sprintf("Houtenae_BG1_ORTH%s.fasta", output))
tools.ExtractSequences(input1, filepath.Join(processDir, input1+".fasta"), input2, filepath.Join(processDir, input2+".fasta"), orthoFileIter1, outFastaIter1)

fmt.Println("2-nd round Alignment ...")
p.runPipelineCmd(processDir, fmt.Sprintf("progressiveMauve --output=%s.vs.Houtenae_BG1_ORTH%s.xmfa --output-guide-tree=%s.vs.Houtenae_BG1_ORTH%s.guide_tree --backbone-output=%s.vs.Houtenae_BG1_ORTH%s.backbone %s.fasta Houtenae_BG1_ORTH%s.fasta", input2, output, input2, output, input2, output, input2, output, logFile), logFile)
p.runPipelineCmd(processDir, fmt.Sprintf("progressiveMauve --output=%s.vs.%s.xmfa --output-guide-tree=%s.vs.%s.guide_tree --backbone-output=%s.vs.%s.backbone %s.fasta %s.fasta", input2, input3, input2, input3, input2, input3, input2, input3, logFile), logFile)

p.runPipelineCmd(processDir, fmt.Sprintf("progBackbonePrep %s.vs.Houtenae_BG1_ORTH%s.backbone >%s.vs.Houtenae_BG1_ORTH%s.backbone.txt", input2, output, input2, output, logFile), logFile)
p.runPipelineCmd(processDir, fmt.Sprintf("progBackbonePrep %s.vs.%s.backbone >%s.vs.%s.backbone.txt", input2, input3, input2, input3, logFile), logFile)

p.runPipelineCmd(processDir, fmt.Sprintf("homBlkReorder %s.vs.Houtenae_BG1_ORTH%s.backbone.txt >%s.vs.Houtenae_BG1_ORTH%s.homblk.txt", input2, output, input2, output, logFile), logFile)
p.runPipelineCmd(processDir, fmt.Sprintf("homBlkReorder %s.vs.%s.backbone.txt >%s.vs.%s.homblk.txt", input2, input3, input2, input3, logFile), logFile)

p.runPipelineCmd(processDir, fmt.Sprintf("orthoParsing %s.vs.Houtenae_BG1_ORTH%s.homblk.txt >%s.vs.Houtenae_BG1_ORTH%s.orthBlk.txt", input2, output, input2, output, logFile), logFile)
p.runPipelineCmd(processDir, fmt.Sprintf("orthoParsing %s.vs.%s.homblk.txt >%s.vs.%s.orthBlk.txt", input2, input3, input2, input3, logFile), logFile)

p.runPipelineCmd(processDir, fmt.Sprintf("patching %s.vs.Houtenae_BG1_ORTH%s.orthBlk.txt %s.vs.%s.orthBlk.txt Houtenae_BG1_ORTH%s.txt %s %s %s_added >Houtenae_BG2_ORTH1%s.txt", input2, output, input2, input3, output, input2, input3, input3, output, logFile), logFile)

fileA2 := filepath.Join(processDir, fmt.Sprintf("%s.vs.Houtenae_BG1_ORTH%s.orthBlk.txt", input2, output))
fileB2 := filepath.Join(processDir, fmt.Sprintf("%s.vs.%s.orthBlk.txt", input2, input3))
fileC2 := filepath.Join(processDir, fmt.Sprintf("Houtenae_BG2_ORTH1%s.txt", output))
fileD2 := filepath.Join(processDir, fmt.Sprintf("Houtenae_BG2_ORTH2%s.txt", output))
tools.InsertIntoFile(fileA2, fileB2, fileC2, fileD2)

fileOut2 := filepath.Join(processDir, fmt.Sprintf("Houtenae_BG2_ORTH%s.txt", output))
tools.ProcessTiaozhen(fileD2, fileOut2, input2, input3, input1)

orthoFileIter2 := filepath.Join(processDir, fmt.Sprintf("Houtenae_BG2_ORTH%s.txt", output))
outFastaIter2 := filepath.Join(processDir, fmt.Sprintf("Houtenae_BG2_ORTH%s.fasta", output))
tools.ExtractSequences(input2, filepath.Join(processDir, input2+".fasta"), input1, filepath.Join(processDir, input1+".fasta"), orthoFileIter2, outFastaIter2)

// NEXT EXTENSION FOR BACTAG2
fmt.Println("1-st round Alignment (Part 2) ...")
p.runPipelineCmd(processDir, fmt.Sprintf("progressiveMauve --output=%s.vs.Houtenae_BG2_ORTH%s.xmfa --output-guide-tree=%s.vs.Houtenae_BG2_ORTH%s.guide_tree --backbone-output=%s.vs.Houtenae_BG2_ORTH%s.backbone %s.fasta Houtenae_BG2_ORTH%s.fasta", input1, output, input1, output, input1, output, input1, output, logFile), logFile)
p.runPipelineCmd(processDir, fmt.Sprintf("progressiveMauve --output=%s.vs.%s.xmfa --output-guide-tree=%s.vs.%s.guide_tree --backbone-output=%s.vs.%s.backbone %s.fasta %s.fasta", input1, input4, input1, input4, input1, input4, input1, input4, logFile), logFile)

p.runPipelineCmd(processDir, fmt.Sprintf("progBackbonePrep %s.vs.Houtenae_BG2_ORTH%s.backbone >%s.vs.Houtenae_BG2_ORTH%s.backbone.txt", input1, output, input1, output, logFile), logFile)
p.runPipelineCmd(processDir, fmt.Sprintf("progBackbonePrep %s.vs.%s.backbone >%s.vs.%s.backbone.txt", input1, input4, input1, input4, logFile), logFile)

p.runPipelineCmd(processDir, fmt.Sprintf("homBlkReorder %s.vs.Houtenae_BG2_ORTH%s.backbone.txt >%s.vs.Houtenae_BG2_ORTH%s.homblk.txt", input1, output, input1, output, logFile), logFile)
p.runPipelineCmd(processDir, fmt.Sprintf("homBlkReorder %s.vs.%s.backbone.txt >%s.vs.%s.homblk.txt", input1, input4, input1, input4, logFile), logFile)

p.runPipelineCmd(processDir, fmt.Sprintf("orthoParsing %s.vs.Houtenae_BG2_ORTH%s.homblk.txt >%s.vs.Houtenae_BG2_ORTH%s.orthBlk.txt", input1, output, input1, output, logFile), logFile)
p.runPipelineCmd(processDir, fmt.Sprintf("orthoParsing %s.vs.%s.homblk.txt >%s.vs.%s.orthBlk.txt", input1, input4, input1, input4, logFile), logFile)

p.runPipelineCmd(processDir, fmt.Sprintf("patching1 %s.vs.Houtenae_BG2_ORTH%s.orthBlk.txt %s.vs.%s.orthBlk.txt Houtenae_BG2_ORTH%s.txt %s %s %s_added >Houtenae_BG3_ORTH1%s.txt", input1, output, input1, input4, output, input1, input4, input4, output, logFile), logFile)

fileA3 := filepath.Join(processDir, fmt.Sprintf("%s.vs.Houtenae_BG2_ORTH%s.orthBlk.txt", input1, output))
fileB3 := filepath.Join(processDir, fmt.Sprintf("%s.vs.%s.orthBlk.txt", input1, input4))
fileC3 := filepath.Join(processDir, fmt.Sprintf("Houtenae_BG3_ORTH1%s.txt", output))
fileD3 := filepath.Join(processDir, fmt.Sprintf("Houtenae_BG3_ORTH2%s.txt", output))
tools.InsertIntoFile(fileA3, fileB3, fileC3, fileD3)

fileOut3 := filepath.Join(processDir, fmt.Sprintf("Houtenae_BG3_ORTH%s.txt", output))
tools.ProcessTiaozhen(fileD3, fileOut3, input1, input4, input2)

orthoFileIter3 := filepath.Join(processDir, fmt.Sprintf("Houtenae_BG3_ORTH%s.txt", output))
outFastaIter3 := filepath.Join(processDir, fmt.Sprintf("Houtenae_BG3_ORTH%s.fasta", output))
tools.ExtractSequences(input1, filepath.Join(processDir, input1+".fasta"), input2, filepath.Join(processDir, input2+".fasta"), orthoFileIter3, outFastaIter3)

fmt.Println("2-nd round Alignment (Part 2) ...")
p.runPipelineCmd(processDir, fmt.Sprintf("progressiveMauve --output=%s.vs.Houtenae_BG3_ORTH%s.xmfa --output-guide-tree=%s.vs.Houtenae_BG3_ORTH%s.guide_tree --backbone-output=%s.vs.Houtenae_BG3_ORTH%s.backbone %s.fasta Houtenae_BG3_ORTH%s.fasta", input2, output, input2, output, input2, output, input2, output, logFile), logFile)
p.runPipelineCmd(processDir, fmt.Sprintf("progressiveMauve --output=%s.vs.%s.xmfa --output-guide-tree=%s.vs.%s.guide_tree --backbone-output=%s.vs.%s.backbone %s.fasta %s.fasta", input2, input4, input2, input4, input2, input4, input2, input4, logFile), logFile)

p.runPipelineCmd(processDir, fmt.Sprintf("progBackbonePrep %s.vs.Houtenae_BG3_ORTH%s.backbone >%s.vs.Houtenae_BG3_ORTH%s.backbone.txt", input2, output, input2, output, logFile), logFile)
p.runPipelineCmd(processDir, fmt.Sprintf("progBackbonePrep %s.vs.%s.backbone >%s.vs.%s.backbone.txt", input2, input4, input2, input4, logFile), logFile)

p.runPipelineCmd(processDir, fmt.Sprintf("homBlkReorder %s.vs.Houtenae_BG3_ORTH%s.backbone.txt >%s.vs.Houtenae_BG3_ORTH%s.homblk.txt", input2, output, input2, output, logFile), logFile)
p.runPipelineCmd(processDir, fmt.Sprintf("homBlkReorder %s.vs.%s.backbone.txt >%s.vs.%s.homblk.txt", input2, input4, input2, input4, logFile), logFile)

p.runPipelineCmd(processDir, fmt.Sprintf("orthoParsing %s.vs.Houtenae_BG3_ORTH%s.homblk.txt >%s.vs.Houtenae_BG3_ORTH%s.orthBlk.txt", input2, output, input2, output, logFile), logFile)
p.runPipelineCmd(processDir, fmt.Sprintf("orthoParsing %s.vs.%s.homblk.txt >%s.vs.%s.orthBlk.txt", input2, input4, input2, input4, logFile), logFile)

p.runPipelineCmd(processDir, fmt.Sprintf("patching1 %s.vs.Houtenae_BG3_ORTH%s.orthBlk.txt %s.vs.%s.orthBlk.txt Houtenae_BG3_ORTH%s.txt %s %s %s_added >Houtenae_BG4_ORTH1%s.txt", input2, output, input2, input4, output, input2, input4, input4, output, logFile), logFile)

fileA4 := filepath.Join(processDir, fmt.Sprintf("%s.vs.Houtenae_BG3_ORTH%s.orthBlk.txt", input2, output))
fileB4 := filepath.Join(processDir, fmt.Sprintf("%s.vs.%s.orthBlk.txt", input2, input4))
fileC4 := filepath.Join(processDir, fmt.Sprintf("Houtenae_BG4_ORTH1%s.txt", output))
fileD4 := filepath.Join(processDir, fmt.Sprintf("Houtenae_BG4_ORTH2%s.txt", output))
tools.InsertIntoFile(fileA4, fileB4, fileC4, fileD4)

fileOut4 := filepath.Join(processDir, fmt.Sprintf("Houtenae_BG4_ORTH%s.txt", output))
tools.ProcessTiaozhen(fileD4, fileOut4, input2, input4, input1)

orthoFileIter4 := filepath.Join(processDir, fmt.Sprintf("Houtenae_BG4_ORTH%s.txt", output))
finalGeneratedFasta := filepath.Join(processDir, fmt.Sprintf("Houtenae_BG4_ORTH%s.fasta", output))
tools.ExtractSequences(input2, filepath.Join(processDir, input2+".fasta"), input1, filepath.Join(processDir, input1+".fasta"), orthoFileIter4, finalGeneratedFasta)

CopyFile(finalGeneratedFasta, filepath.Join(p.OutputDir, output+".fasta"))

return nil
}

// CopyFile is a helper utility
func CopyFile(src, dst string) error {
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
