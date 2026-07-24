package tools

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// ExtractSequences performs the function of seqExtFromMultiGenome.pl
func ExtractSequences(genome1File, genome1Fasta, genome2File, genome2Fasta, positionsFile, outputFile string) error {
	genomes := make(map[string]string)

	readFasta := func(prefix string, filepath string) error {
		file, err := os.Open(filepath)
		if err != nil {
			return err
		}
		defer file.Close()

		var seqBuilder strings.Builder
		scanner := bufio.NewScanner(file)
		// increase max scanner capacity for long sequences
		buf := make([]byte, 0, 1024*1024)
		scanner.Buffer(buf, 100*1024*1024)

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, ">") {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) > 0 {
				seqBuilder.WriteString(fields[0])
			}
		}
		genomes[prefix] = seqBuilder.String()
		return scanner.Err()
	}

	if err := readFasta(genome1File, genome1Fasta); err != nil {
		return err
	}
	if err := readFasta(genome2File, genome2Fasta); err != nil {
		return err
	}

	pos, err := os.Open(positionsFile)
	if err != nil {
		return err
	}
	defer pos.Close()

	var agBuilder strings.Builder
	scanner := bufio.NewScanner(pos)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			sp := fields[0]
			st, err1 := strconv.Atoi(fields[1])
			en, err2 := strconv.Atoi(fields[2])

			if err1 == nil && err2 == nil {
				seqLen := en - st + 1
				stIdx := st - 1

				if seq, exists := genomes[sp]; exists {
					if stIdx >= 0 && stIdx+seqLen <= len(seq) {
						agBuilder.WriteString(seq[stIdx : stIdx+seqLen])
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	out, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer out.Close()
	writer := bufio.NewWriter(out)

	writer.WriteString(">Salmonella orthologous_ancient genome\n")
	agStr := agBuilder.String()
	for i := 0; i < len(agStr); i += 80 {
		end := i + 80
		if end > len(agStr) {
			end = len(agStr)
		}
		writer.WriteString(agStr[i:end] + "\n")
	}

	return writer.Flush()
}
