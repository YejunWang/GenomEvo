/* Usage: blastFilter Blast_Out_Flie Threshold_Of_SIMILARITY_And_Coverage
   Example: blastFilter blastout.txt 0.7 */
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

func blastFilterApp(args []string) {
	if len(args) < 3 {
		fmt.Printf("No input file or parameter!\n")
		return
	}

	inputFileName := args[1]
	simThreshold, _ := strconv.ParseFloat(args[2], 16)
	batchNum := strings.Split(inputFileName, "/")[len(strings.Split(inputFileName, "/"))-1]
	batchNum = strings.TrimSuffix(batchNum, "_blastout.txt")
	inputFile, openErr := os.Open(inputFileName)
	if openErr != nil {
		fmt.Printf("Error happened in opening %s!\n", inputFileName)
		return
	}
	defer inputFile.Close()

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

	// 获取该batch相应菌株的蛋白序列长度信息
	strseqLen := make(map[string]float64)
	for _, strain := range strains {
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
				strseqLen[lineArray[0]], _ = strconv.ParseFloat(lineArray[2], 16)
			}
			if readErr == io.EOF {
				break
			}
		}
	}

	// 筛选blast比对结果
	inputReader := bufio.NewReader(inputFile)
	type maxRow struct {
		maxSubject string
		maxValue   []float64
	}
	queryMap := make(map[string]map[int]maxRow)
	skipMap := make(map[string]string)
	var query0 string
	var queryNum, subjectNum int
	for {
		line, readErr := inputReader.ReadString('\n')
		line = strings.Trim(line, "\r\n")
		line = strings.Trim(line, "\n")
		if len(line) != 0 {
			lineArray := strings.Split(line, "\t")
			query := lineArray[0]
			subject := lineArray[1]
			queryId := strings.Split(query, "|-|")[0]
			subjectId := strings.Split(subject, "|-|")[0]
			queryNum, _ = strconv.Atoi(strings.Split(query, "|-|")[1])
			subjectNum, _ = strconv.Atoi(strings.Split(subject, "|-|")[1])
			sim, _ := strconv.ParseFloat(lineArray[2], 16)
			covLen, _ := strconv.ParseFloat(lineArray[3], 16)
			// rem := strings.Join(lineArray[4:], "\t")

			// 判断是否为新的query id，若为新query，则将上个query结果输出，若仍为旧query，继续筛选/比较
			if query != query0 {
				if len(query0) != 0 {
					// 将上一条query蛋白的结果输出
					// 同时过滤在此输出的subject的同名query蛋白
					// 以保证不会存在query1-subject1和subject1-query1同时输出的情况
					for query, subjects := range queryMap {
						for _, data := range subjects {
							skipMap[data.maxSubject] = "skip"
							fmt.Printf("%s\t%s\t%f\t%f\n", query, data.maxSubject, data.maxValue[0], data.maxValue[1])
						}
					}
				}
				query0 = query
				queryMap = make(map[string]map[int]maxRow)
			}
			if queryNum >= subjectNum {
				continue
			} else if skipMap[query] == "skip" || skipMap[subject] == "skip" {
				// 到底应该是 仅有一个 skipMap[query] == "skip"
				// 还是应该是 skipMap[query] == "skip" || skipMap[subject] == "skip"
				// 记录：仅有一个 skipMap[query] == "skip" 时，会出现 query1-subject1 和 query2-subject1 同时输出的情况
				// 原因是，比对结果文件出现了 SeqA-SeqB 相似度和或覆盖度不达标的情况，但是 SeqA-SeqC 和 SeqB-SeqC 同时达标
				// 应该把 SeqC 算在哪个家族中呢？还是两个都算呢？
				// 记得同步修改 iterblastFilter.go 中的代码
				continue
			} else {
				cov := ((covLen / strseqLen[queryId]) + (covLen / strseqLen[subjectId])) / 2
				if cov < simThreshold || (sim/100) < simThreshold {
					continue
				} else {
					if _, ok := queryMap[query]; !ok {
						queryMap[query] = make(map[int]maxRow)
					}
					if _, ok := queryMap[query][subjectNum]; !ok {
						queryMap[query][subjectNum] = maxRow{}
					}

					product := sim * covLen

					if queryMap[query][subjectNum].maxSubject == "" {
						// 第一次遇到该 subject，直接记录
						queryMap[query][subjectNum] = maxRow{maxSubject: subject, maxValue: []float64{sim, covLen}}
					} else {
						// 如果已经记录过该 subject，更新乘积和覆盖度
						if product > queryMap[query][subjectNum].maxValue[0]*queryMap[query][subjectNum].maxValue[1] {
							queryMap[query][subjectNum] = maxRow{maxSubject: subject, maxValue: []float64{sim, covLen}}
						}
					}
				}
			}
		}
		if readErr == io.EOF {
			// 将最后一条query蛋白的结果输出
			for query, subjects := range queryMap {
				for _, data := range subjects {
					skipMap[data.maxSubject] = "skip"
					fmt.Printf("%s\t%s\t%f\t%f\n", query, data.maxSubject, data.maxValue[0], data.maxValue[1])
				}
			}
			break
		}
	}
}
