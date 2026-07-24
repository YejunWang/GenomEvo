package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

func iterPGtmp2fastaApp(args []string) {
	if len(args) < 2 {
		fmt.Printf("Lack of input file!!\n")
		return
	}

	pgtmpFileName := args[1]
	pgtmpFile, openErr := os.Open(pgtmpFileName)
	if openErr != nil {
		log.Printf("Error happened in opening %s: %v", pgtmpFileName, openErr)
		return
	}
	defer pgtmpFile.Close()

	iterNumstr := strings.Split(pgtmpFileName, "/")[len(strings.Split(pgtmpFileName, "/"))-1]
	iterNumstr = strings.TrimPrefix(iterNumstr, "iter")
	iterNumstr = strings.TrimSuffix(iterNumstr, ".PG.tmp1.txt")
	iterNum, _ := strconv.Atoi(iterNumstr)
	seqnameSuffix := "|-|I" + iterNumstr

	var iterOrbatch []string
	if iterNum == 2 {
		iterOrbatch = []string{"./result/batch/batch1/batch1.PG.tmp.fasta", "./result/batch/batch2/batch2.PG.tmp.fasta"}
	} else {
		preiterNum := iterNum - 1
		iterOrbatch = []string{"./result/iter/iter" + strconv.Itoa(preiterNum) + "/iter" + strconv.Itoa(preiterNum) + ".PG.tmp.fasta", "./result/batch/batch" + strconv.Itoa(iterNum) + "/batch" + strconv.Itoa(iterNum) + ".PG.tmp.fasta"}
	}

	pgtmpReader := bufio.NewReader(pgtmpFile)
	// 获取需要输出到fasta文件的序列名
	tmpfastaMap := make(map[string]string)
	for {
		line, readErr := pgtmpReader.ReadString('\n')
		line = strings.Trim(line, "\r\n")
		line = strings.Trim(line, "\n")
		if len(line) != 0 {
			if strings.HasPrefix(line, "PG_b") || strings.HasPrefix(line, "PG_iter") {
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

	// 遍历该iter相应上一个iter或batch的pgtmpfasta，若其存在于tmpfastaMap中，则将该蛋白序列输出
	var output bool
	for _, v := range iterOrbatch {
		pgtmpfastaFile, openErr := os.Open(v)
		if openErr != nil {
			fmt.Printf("Error happened in opening %s!\n", v)
			return
		}
		defer pgtmpfastaFile.Close()

		pgtmpfastaReader := bufio.NewReader(pgtmpfastaFile)
		for {
			line, readErr := pgtmpfastaReader.ReadString('\n')
			line = strings.Trim(line, "\r\n")
			line = strings.Trim(line, "\n")
			if len(line) != 0 {
				if strings.HasPrefix(line, ">") {
					seqId := strings.TrimPrefix(line, ">")
					if strings.Contains(seqId, "|-|I") {
						seqId = strings.Split(seqId, "|-|I")[0]
					} else if strings.Contains(seqId, "|-|B") {
						seqId = strings.Split(seqId, "|-|B")[0]
					}
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
