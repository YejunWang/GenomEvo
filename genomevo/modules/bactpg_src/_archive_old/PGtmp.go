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

func PGtmpApp(args []string) {
	if len(args) < 2 {
		fmt.Printf("Lack of input file!!\n")
		return
	}

	blafilterFileName := args[1]
	batchNum := strings.Split(blafilterFileName, "/")[len(strings.Split(blafilterFileName, "/"))-1]
	batchNum = strings.TrimSuffix(batchNum, "_blastoutFilter.txt")
	midColname := "b" + strings.Split(batchNum, "batch")[1] + "M"
	blafilterFile, openErr := os.Open(blafilterFileName)
	if openErr != nil {
		fmt.Printf("Error happened in opening %s!\n", blafilterFileName)
		return
	}
	defer blafilterFile.Close()

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

	// 该batch的菌株数
	strainsNum := len(strains)

	// 打印列名
	var colName []string
	for i := 1; i <= strainsNum; i++ {
		stri := strconv.Itoa(i)
		colName = append(colName, "PG_"+midColname+"_"+stri)
	}
	fmt.Print(strings.Join(colName, "\t"), "\n")

	blafilterReader := bufio.NewReader(blafilterFile)
	var query0 string
	homoseqMap := make(map[int]string)
	homoseqRow := []string{}
	exitseqMap := make(map[string]string)
	for {
		line, readErr := blafilterReader.ReadString('\n')
		line = strings.Trim(line, "\r\n")
		line = strings.Trim(line, "\n")
		if len(line) != 0 {
			lineArray := strings.Split(line, "\t")
			query := lineArray[0]
			subject := lineArray[1]
			queryId := strings.Split(query, "|-|")[0]
			subjectId := strings.Split(subject, "|-|")[0]
			queryNum, _ := strconv.Atoi(strings.Split(query, "|-|")[1])
			subjectNum, _ := strconv.Atoi(strings.Split(subject, "|-|")[1])

			if query != query0 {
				exitseqMap[queryId] = "exit"
				if len(query0) != 0 {
					// 将上一条query蛋白的结果输出
					for i := 1; i <= strainsNum; i++ {
						if len(homoseqMap[i]) == 0 {
							homoseqMap[i] = "-"
							homoseqRow = append(homoseqRow, homoseqMap[i])
						} else {
							homoseqRow = append(homoseqRow, homoseqMap[i])
						}
					}
					fmt.Print(strings.Join(homoseqRow, "\t"), "\n")
					homoseqMap = map[int]string{}
					homoseqRow = []string{}
				}
				query0 = query
				homoseqMap[queryNum] = queryId
			}
			homoseqMap[subjectNum] = subjectId
			exitseqMap[subjectId] = "exit"
		}
		if readErr == io.EOF {
			// 将最后一条query蛋白的结果输出
			for i := 1; i <= strainsNum; i++ {
				if len(homoseqMap[i]) == 0 {
					homoseqMap[i] = "-"
					homoseqRow = append(homoseqRow, homoseqMap[i])
				} else {
					homoseqRow = append(homoseqRow, homoseqMap[i])
				}
			}
			fmt.Print(strings.Join(homoseqRow, "\t"), "\n")
			homoseqMap = map[int]string{}
			homoseqRow = []string{}
			break
		}
	}

	// 获取该batch相应菌株的各家族代表蛋白，若其未存在于以上输出结果中，则将该特异蛋白输出
	for k, strain := range strains {
		strainOrder := k + 1
		strmapFileName := "./result/str_famap/" + strain + ".map"
		strmapFile, openErr := os.Open(strmapFileName)
		if openErr != nil {
			fmt.Printf("Error happened in opening %s!\n", strmapFileName)
			return
		}
		defer strmapFile.Close()

		strmapReader := bufio.NewReader(strmapFile)
		for {
			line, readErr := strmapReader.ReadString('\n')
			line = strings.Trim(line, "\r\n")
			line = strings.Trim(line, "\n")
			if len(line) != 0 {
				lineArray := strings.Split(line, "\t")
				seqId := lineArray[0]
				if exitseqMap[seqId] == "exit" {
					continue
				} else {
					homoseqMap[strainOrder] = seqId
					for i := 1; i <= strainsNum; i++ {
						if len(homoseqMap[i]) == 0 {
							homoseqMap[i] = "-"
							homoseqRow = append(homoseqRow, homoseqMap[i])
						} else {
							homoseqRow = append(homoseqRow, homoseqMap[i])
						}
					}
					fmt.Print(strings.Join(homoseqRow, "\t"), "\n")
					homoseqMap = map[int]string{}
					homoseqRow = []string{}
				}
			}
			if readErr == io.EOF {
				break
			}
		}
	}
}
