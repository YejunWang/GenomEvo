package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func PGtmp2fastaApp(args []string) {
	if len(args) < 2 {
		fmt.Printf("Lack of input file!!\n")
		return
	}

	pgtmpFileName := args[1]
	batchNum := strings.Split(pgtmpFileName, "/")[len(strings.Split(pgtmpFileName, "/"))-1]
	batchNum = strings.TrimSuffix(batchNum, ".PG.tmp1.txt")
	seqnameSuffix := "|-|B" + strings.Split(batchNum, "batch")[1]

	pgtmpFile, openErr := os.Open(pgtmpFileName)
	if openErr != nil {
		fmt.Printf("Error happened in opening %s!\n", pgtmpFileName)
		return
	}
	defer pgtmpFile.Close()

	// 获取该batch相应菌株
	strains := []string{}
	supFile, supopenErr := os.Open("./result/superParamer.txt")
	if supopenErr != nil {
		log.Fatal(supopenErr)
	}
	defer supFile.Close()

	scanner := bufio.NewScanner(supFile)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.Trim(line, "\r\n")
		line = strings.Trim(line, "\n")
		lineBlocks := strings.Split(line, "\t")
		if lineBlocks[0] == batchNum {
			strains = append(strains, lineBlocks[1])
		}
	}

	pgtmpReader := bufio.NewReader(pgtmpFile)
	// 获取需要输出到fasta文件的序列名
	tmpfastaMap := make(map[string]string)
	for {
		line, readErr := pgtmpReader.ReadString('\n')
		line = strings.Trim(line, "\r\n")
		line = strings.Trim(line, "\n")
		if len(line) != 0 {
			if strings.HasPrefix(line, "PG_b") {
				continue
			} else {
				lineArray := strings.Split(line, "\t")
				for _, seqId := range lineArray {
					if seqId == "-" {
						continue
					} else {
						tmpfastaMap[seqId] = "output"
						break
					}
				}
			}
		}
		if readErr == io.EOF {
			break
		}
	}

	// 遍历该batch相应菌株的nr_seq，若其存在于tmpfastaMap中，则将该蛋白序列输出
	var output bool
	for _, strain := range strains {
		nrseqFileName := "./result/nr_seq/" + strain + ".fasta"
		nrseqFile, openErr := os.Open(nrseqFileName)
		if openErr != nil {
			fmt.Printf("Error happened in opening %s!\n", nrseqFileName)
			return
		}
		defer nrseqFile.Close()

		nrseqReader := bufio.NewReader(nrseqFile)
		for {
			line, readErr := nrseqReader.ReadString('\n')
			line = strings.Trim(line, "\r\n")
			line = strings.Trim(line, "\n")
			if len(line) != 0 {
				if strings.HasPrefix(line, ">") {
					seqId := strings.TrimPrefix(line, ">")
					if tmpfastaMap[seqId] == "output" {
						fmt.Print(">"+seqId+seqnameSuffix, "\n")
						output = true
					} else {
						output = false
					}
				} else if output == true {
					fmt.Println(line)
				}
			}
			if readErr == io.EOF {
				break
			}
		}
	}
}
