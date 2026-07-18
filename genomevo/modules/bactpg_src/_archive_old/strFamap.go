package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func strFamapApp(args []string) {
	if len(args) <= 2 {
		fmt.Printf("Lack of one or more input files!\n")
		return
	}

	seqFileName := args[1]
	seqFile, seqopenErr := os.Open(seqFileName)
	if seqopenErr != nil {
		fmt.Printf("Error happened in opening %s!\n", seqFileName)
		return
	}
	defer seqFile.Close()

	clstrFileName := args[2]
	clstrFile, clstropenErr := os.Open(clstrFileName)
	if clstropenErr != nil {
		fmt.Printf("Error happened in opening %s!\n", clstrFileName)
		return
	}
	defer clstrFile.Close()

	seqReader := bufio.NewReader(seqFile)
	lenMap := make(map[string]int)
	var gene, gene0, seq string
	for {
		line, readErr := seqReader.ReadString('\n')
		line = strings.Trim(line, "\r\n")
		line = strings.Trim(line, "\n")
		if strings.Contains(line, ">") {
			line = strings.Replace(line, ">", "", -1)
			lineArray1 := strings.Fields(line)
			gene0 = lineArray1[0]
			if len(gene) > 0 {
				lenMap[gene] = len(seq)
				seq = ""
			}
			gene = gene0
		} else if len(line) != 0 {
			seq += line
		}
		if readErr == io.EOF {
			lenMap[gene] = len(seq)
			break
		}
	}

	clstrReader := bufio.NewReader(clstrFile)
	var repSeq, homoSeq, homoSeqs string
	for {
		line, readErr := clstrReader.ReadString('\n')
		line = strings.Trim(line, "\r\n")
		line = strings.Trim(line, "\n")
		if strings.HasPrefix(line, ">") {
			if len(repSeq) != 0 && len(homoSeqs) != 0 {
				fmt.Printf("%s\t%s\t%d\n", repSeq, repSeq+"|"+homoSeqs, lenMap[repSeq])
			} else if len(repSeq) != 0 && len(homoSeqs) == 0 {
				fmt.Printf("%s\t%s\t%d\n", repSeq, repSeq, lenMap[repSeq])
			}
			repSeq = ""
			homoSeq = ""
			homoSeqs = ""
		} else if strings.Contains(line, "aa") && strings.Contains(line, ">") && strings.Contains(line, "*") {
			repSeq = strings.Replace(line, "\t", "", -1)
			repSeq = strings.Split(repSeq, " ")[1]
			repSeq = strings.Replace(repSeq, ">", "", -1)
			repSeq = strings.TrimRight(repSeq, ".")
		} else if strings.Contains(line, "aa") && strings.Contains(line, ">") && !strings.Contains(line, "*") {
			homoSeq = strings.Replace(line, "\t", "", -1)
			homoSeq = strings.Fields(homoSeq)[1]
			homoSeq = strings.Replace(homoSeq, ">", "", -1)
			homoSeq = strings.TrimRight(homoSeq, ".")
			if len(homoSeqs) != 0 {
				homoSeqs = homoSeqs + "|" + homoSeq
			} else {
				homoSeqs = homoSeq
			}
		}
		if readErr == io.EOF {
			if len(repSeq) != 0 && len(homoSeqs) != 0 {
				fmt.Printf("%s\t%s\t%d\n", repSeq, homoSeqs, lenMap[repSeq])
			} else if len(repSeq) != 0 && len(homoSeqs) == 0 {
				fmt.Printf("%s\t%s\t%d\n", repSeq, repSeq, lenMap[repSeq])
			}
			break
		}
	}
}
