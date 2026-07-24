package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var clustalDir string
var clustalGcgDir string
var clustalMegaDir string

var clustalCmd = &cobra.Command{
	Use:   "clustal",
	Short: "Run clustal sequence alignment and convert to MEGA format",
	Long:  `Port of Tool/clustal.pl and Tool/gcg2meg.pl. Aligns .fa files in the specified directory using clustalw2 and saves the resulting mega files.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := runClustal(clustalDir)
		if err != nil {
			fmt.Println("Error running clustal:", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(clustalCmd)
	clustalCmd.Flags().StringVarP(&clustalDir, "dir", "d", "output/CG_results/all-strain-together/2.result", "Directory containing .fa files to process")
	clustalCmd.Flags().StringVar(&clustalGcgDir, "gcgdir", "", "Output directory for GCG files (default: <dir>/../3.gcg)")
	clustalCmd.Flags().StringVar(&clustalMegaDir, "megadir", "", "Output directory for MEGA files (default: <dir>/../3.mega)")
}

func runClustal(dir string) error {
	// Determine output directories
	gcgDir := clustalGcgDir
	megaDir := clustalMegaDir
	if gcgDir == "" {
		gcgDir = filepath.Join(filepath.Dir(dir), "3.gcg")
	}
	if megaDir == "" {
		megaDir = filepath.Join(filepath.Dir(dir), "3.mega")
	}
    
	if err := os.MkdirAll(gcgDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(megaDir, 0755); err != nil {
		return err
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".fa") {
			fileName := filepath.Join(dir, f.Name())
			filePrefix := strings.TrimSuffix(f.Name(), ".fa")
			outFile := filepath.Join(gcgDir, filePrefix+".gcg")
			megaFile := filepath.Join(megaDir, filePrefix+".meg")

			// Check for clustalw2 executable path
			clustalPath := "Tool/clustalw2"
			if path, err := exec.LookPath("clustalw2"); err == nil {
				clustalPath = path
			}

			// Run clustalw2
			fmt.Printf("Aligning %s -> %s\n", fileName, outFile)
			cmd := exec.Command(clustalPath,
				"-infile="+fileName,
				"-type=protein",
				"-output=gcg",
				"-outfile="+outFile,
				"-pwmatrix=GONNET",
				"-pwgapopen=10",
				"-pwgapext=0.1",
				"-gapopen=10",
				"-gapext=0.2",
				"-gapdist=4",
				"-align")

			if out, err := cmd.CombinedOutput(); err != nil {
				fmt.Printf("Clustalw2 failed for %s: %s\n", fileName, out)
				// Might continue to next file
                continue
			}

			// Convert to MEGA format
			fmt.Printf("Converting %s -> %s\n", outFile, megaFile)
			if err := gcg2Meg(outFile, megaFile); err != nil {
				fmt.Printf("Conversion to meg failed for %s: %v\n", outFile, err)
			}
		}
	}
	return nil
}

func gcg2Meg(infile, outfile string) error {
	in, err := os.Open(infile)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(outfile)
	if err != nil {
		return err
	}
	defer out.Close()

	seqMap := make(map[string]string)
	flag := false
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "//") {
			flag = true
			continue
		}
		if flag {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				// The first token is the strain, the rest is sequence
				strain := parts[0]
				seqPart := strings.Join(parts[1:], "")
				seqPart = strings.ReplaceAll(seqPart, " ", "")
				seqMap[strain] = seqMap[strain] + seqPart
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// Write MEGA format header
	writer := bufio.NewWriter(out)
	writer.WriteString("#mega\n!Title CG;\n!Format DataType=Protein indel=-;\n\n")

	for strain, sequence := range seqMap {
		writer.WriteString("#" + strain + "\n")
		// Replace '.' with '-'
		seqFormatted := strings.ReplaceAll(sequence, ".", "-")
		writer.WriteString(seqFormatted + "\n\n")
	}

	return writer.Flush()
}
