package main
import (
    "bufio"
    "fmt"
    "io"
    "log"
    "os"
    "os/exec"
    "sort"
    "strconv"
    "strings"
)
func iterPGtmpApp(args []string) {
	if len(args) < 2 {
		fmt.Printf("Lack of input file!!\n")
		return
	}

	blafilterFileName := args[1]
	iterNumstr := strings.Split(blafilterFileName, "/")[len(strings.Split(blafilterFileName, "/"))-1]
	iterNumstr = strings.TrimPrefix(iterNumstr, "iter")
	iterNumstr = strings.TrimSuffix(iterNumstr, "_blastoutFilter.txt")
	iterNum, _ := strconv.Atoi(iterNumstr)

	var iterOrbatch []string
	if iterNum == 2 {
		iterOrbatch = []string{
			"./result/batch/batch1/batch1.PG.tmp1.txt",
			"./result/batch/batch2/batch2.PG.tmp1.txt"}
	} else {
		iterOrbatch = []string{
			"./result/iter/iter" + strconv.Itoa(iterNum-1) + "/iter" + strconv.Itoa(iterNum-1) + ".PG.tmp1.txt",
			"./result/batch/batch" + strconv.Itoa(iterNum) + "/batch" + strconv.Itoa(iterNum) + ".PG.tmp1.txt"}
	}

	// 获取前两个文件的表头及各自的列数
	cmd := exec.Command("head", "-n", "1", iterOrbatch[0])
	output, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
		return
	}
	colName0 := strings.Trim(string(output), "\r\n")
	colName0 = strings.Trim(colName0, "\n")
	colNum0 := len(strings.Split(colName0, "\t"))

	cmd = exec.Command("head", "-n", "1", iterOrbatch[1])
	output, err = cmd.Output()
	if err != nil {
		fmt.Println(err)
		return
	}
	colName1 := strings.Trim(string(output), "\r\n")
	colName1 = strings.Trim(colName1, "\n")
	colNum1 := len(strings.Split(colName1, "\t"))

	// 打印列名
	fmt.Print(colName0 + "\t" + colName1 + "\n")

	// 创建一个 map 存储同源蛋白家族蛋白名，便于后续输出最终结果：不存在同源家族蛋白，独立输出，存在则输出至同一行
	exitseqMap := make(map[string]string)
	// 创建一个 map 存储相互配对的同源家族蛋白名
	pairMap := make(map[string][]string)

	blafilterFile, openErr := os.Open(blafilterFileName)
	if openErr != nil {
		fmt.Printf("Error happened in opening %s!\n", blafilterFileName)
		return
	}
	defer blafilterFile.Close()
	blafilterReader := bufio.NewReader(blafilterFile)
	var queryId, subjectId string
	for {
		line, readErr := blafilterReader.ReadString('\n')
		line = strings.Trim(line, "\r\n")
		line = strings.Trim(line, "\n")
		if len(line) != 0 {
			lineArray := strings.Split(line, "\t")
			query := lineArray[0]
			subject := lineArray[1]
			if strings.Contains(query, "|-|I") {
				queryId = strings.Split(query, "|-|I")[0]
			} else if strings.Contains(query, "|-|B") {
				queryId = strings.Split(query, "|-|B")[0]
			}
			if strings.Contains(subject, "|-|I") {
				subjectId = strings.Split(subject, "|-|I")[0]
			} else if strings.Contains(subject, "|-|B") {
				subjectId = strings.Split(subject, "|-|B")[0]
			}

			exitseqMap[queryId] = "exit"
			exitseqMap[subjectId] = "exit"
			pairMap[subjectId] = append(pairMap[subjectId], queryId)
		}
		if readErr == io.EOF {
			break
		}
	}

	// 获取该iter相应的batch/iter的各家族代表蛋白，存在同源家族蛋白输出至同一行，不存在则输出至独立一行
	firstpgMap := make(map[string]string)
	for i, v := range iterOrbatch {
		pgtmpFile, opentmpErr := os.Open(v)
		if opentmpErr != nil {
			fmt.Printf("Error happened in opening %s!\n", v)
			return
		}
		defer pgtmpFile.Close()
		pgtmpReader := bufio.NewReader(pgtmpFile)

		if i == 0 {
			for {
				line, readErr := pgtmpReader.ReadString('\n')
				line = strings.Trim(line, "\r\n")
				line = strings.Trim(line, "\n")
				// 跳过首行列名或空行
				if len(line) != 0 && !strings.Contains(line, "PG_b") && !strings.Contains(line, "PG_iter") {
					lineArray := strings.Split(line, "\t")
					for _, seqId := range lineArray {
						if seqId == "-" {
							continue
						} else {
							if exitseqMap[seqId] == "exit" {
								firstpgMap[seqId] = line
								break
							} else if len(exitseqMap[seqId]) == 0 {
								// 将 - 重复指定次数
								str := strings.Repeat("-", colNum1)
								// 使用 SplitN 函数将字符串分割成切片，n 参数为切片中元素的数量
								strSlice := strings.SplitAfterN(str, "-", colNum1)
								strOut := strings.Join(strSlice, "\t")
								fmt.Println(line + "\t" + strOut)
								break
							}
						}
					}

				}
				if readErr == io.EOF {
					break
				}
			}
		} else if i == 1 {
			for {
				line, readErr := pgtmpReader.ReadString('\n')
				line = strings.Trim(line, "\r\n")
				line = strings.Trim(line, "\n")
				// 跳过首行列名或空行
				if len(line) != 0 && !strings.Contains(line, "PG_b") && !strings.Contains(line, "PG_iter") {
					lineArray := strings.Split(line, "\t")
					for _, seqId := range lineArray {
						if seqId == "-" {
							continue
						} else {
							if exitseqMap[seqId] == "exit" {
								for _, v := range pairMap[seqId] {
									fmt.Println(firstpgMap[v] + "\t" + line)
								}
								break
							} else if len(exitseqMap[seqId]) == 0 {
								// 将 - 重复指定次数
								str := strings.Repeat("-", colNum0)
								// 使用 SplitN 函数将字符串分割成切片，n 参数为切片中元素的数量
								strSlice := strings.SplitAfterN(str, "-", colNum0)
								strOut := strings.Join(strSlice, "\t")
								fmt.Println(strOut + "\t" + line)
								break
							}
						}
					}

				}
				if readErr == io.EOF {
					break
				}
			}
		}
	}
}

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

func seqCountApp(args []string) {
	if len(args) < 2 {
		fmt.Printf("Lack of input file!\n")
		return
	}

	fileName := args[1]
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	count := 0
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, ">") {
			count++
		}
	}

	fmt.Print(count)
}

func batchseqMergeApp(args []string) {
	if len(args) <= 2 {
		fmt.Printf("Lack of one or more input parameters!\n")
		return
	}

	seqFileName := args[1]
	orderNum := args[2]
	seqFile, seqopenErr := os.Open(seqFileName)
	if seqopenErr != nil {
		fmt.Printf("Error happened in opening %s!\n", seqFileName)
		return
	}
	defer seqFile.Close()

	seqReader := bufio.NewReader(seqFile)
	for {
		line, readErr := seqReader.ReadString('\n')
		line = strings.Trim(line, "\r\n")
		line = strings.Trim(line, "\n")
		if strings.Contains(line, ">") {
			line = line + "|-|" + orderNum
			fmt.Print(line, "\n")
		} else if len(line) != 0 {
			fmt.Print(line, "\n")
		}
		if readErr == io.EOF {
			fmt.Print(line, "\n")
			break
		}
	}
}

func iterseqMergeApp(args []string) {
	if len(args) < 2 {
		fmt.Printf("Lack of input file!\n")
		return
	}

	seqFileName := args[1]
	seqFile, seqopenErr := os.Open(seqFileName)
	if seqopenErr != nil {
		fmt.Printf("Error happened in opening %s!\n", seqFileName)
		return
	}
	defer seqFile.Close()

	seqReader := bufio.NewReader(seqFile)
	for {
		line, readErr := seqReader.ReadString('\n')
		line = strings.Trim(line, "\r\n")
		line = strings.Trim(line, "\n")
		if strings.Contains(line, ">") {
			fmt.Print(line, "\n")
		} else if len(line) != 0 {
			fmt.Print(line, "\n")
		}
		if readErr == io.EOF {
			fmt.Print(line, "\n")
			break
		}
	}
}

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

func iterblastFilterApp(args []string) {
	if len(args) < 3 {
		fmt.Printf("No input file or parameter!\n")
		return
	}
	inputFileName := args[1]
	simThreshold, _ := strconv.ParseFloat(args[2], 16)
	iterNumstr := strings.Split(inputFileName, "/")[len(strings.Split(inputFileName, "/"))-1]
	iterNumstr = strings.TrimPrefix(iterNumstr, "iter")
	iterNumstr = strings.TrimSuffix(iterNumstr, "_blastout.txt")
	iterNum, _ := strconv.Atoi(iterNumstr)

	inputFile, openErr := os.Open(inputFileName)
	if openErr != nil {
		fmt.Printf("Error happened in opening %s!\n", inputFileName)
		return
	}
	defer inputFile.Close()

	// 获取该iter相应菌株
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
		batchNum, _ := strconv.Atoi(strings.TrimPrefix(lineBlocks[0], "batch"))
		if batchNum <= iterNum {
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
	var query0, queryId, subjectId string
	var queryNum, subjectNum int
	for {
		line, readErr := inputReader.ReadString('\n')
		line = strings.Trim(line, "\r\n")
		line = strings.Trim(line, "\n")
		if len(line) != 0 {
			lineArray := strings.Split(line, "\t")
			query := lineArray[0]
			subject := lineArray[1]
			// 对于batch内的序列，seqId后面跟的是下划线加数字如'|-|1'
			// 对于batch间的序列，seqId后面跟的是|-|加B加数字如'|-|B1'
			// 对于iter序列，seqId后面跟的是|-|加I加数字如'|-|I1'
			if strings.Contains(query, "|-|I") {
				queryId = strings.Split(query, "|-|I")[0]
				queryNum, _ = strconv.Atoi(strings.Split(query, "|-|I")[1])
			} else if strings.Contains(query, "|-|B") {
				queryId = strings.Split(query, "|-|B")[0]
				queryNum, _ = strconv.Atoi(strings.Split(query, "|-|B")[1])
			}
			if strings.Contains(subject, "|-|I") {
				subjectId = strings.Split(subject, "|-|I")[0]
				subjectNum, _ = strconv.Atoi(strings.Split(subject, "|-|I")[1])
			} else if strings.Contains(subject, "|-|B") {
				subjectId = strings.Split(subject, "|-|B")[0]
				subjectNum, _ = strconv.Atoi(strings.Split(subject, "|-|B")[1])
			}
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
				// 多种情况讨论，详见 blastFilter.go 中注释说明
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

func chang_seqid_pgidApp(args []string) {
	if len(args) < 2 {
		fmt.Printf("No input file!\n")
		return
	}

	lastIterName := args[1]

	// 读取所有菌株的同菌株同源家族蛋白名，生成map
	strseqMap := make(map[string]string)
	str_famap := "./result/str_famap"
	str_famapList, str_famapErr := os.ReadDir(str_famap)
	if str_famapErr != nil {
		fmt.Printf("Error happened in the iterfile listing step!\n")
		return
	}

	for _, fi := range str_famapList {
		str_famapFile, openErr := os.Open("./result/str_famap/" + fi.Name())
		if openErr != nil {
			fmt.Printf("Error happened in opening %s!\n", fi.Name())
			return
		}
		defer str_famapFile.Close()
		strmapReader := bufio.NewReader(str_famapFile)
		for {
			line, readErr := strmapReader.ReadString('\n')
			line = strings.Trim(line, "\r\n")
			line = strings.Trim(line, "\n")
			if len(line) != 0 {
				lineArray := strings.Split(line, "\t")
				if lineArray[0] != lineArray[1] {
					strseqMap[lineArray[0]] = lineArray[1]
				}
			}
			if readErr == io.EOF {
				break
			}
		}
	}

	// 打印表头/列名
	var colNames []string
	supDir := "./result/superParamer.txt"
	supFile, supOpenerr := os.Open(supDir)
	if supOpenerr != nil {
		fmt.Printf("Error happened in opening %s!\n", supDir)
		return
	}
	defer supFile.Close()
	supReader := bufio.NewReader(supFile)
	for {
		line, readErr := supReader.ReadString('\n')
		line = strings.Trim(line, "\r\n")
		line = strings.Trim(line, "\n")
		if len(line) != 0 {
			lineBlocks := strings.Split(line, "\t")
			colNames = append(colNames, lineBlocks[1])
		}
		if readErr == io.EOF {
			break
		}
	}
	fmt.Println("PG_ID" + "\t" + strings.Join(colNames, "\t"))

	// 添加PG_ID，以及替换同菌株同源蛋白序列名称
	lastIterfasta, openErr := os.Open(lastIterName)
	if openErr != nil {
		fmt.Printf("Error happened in opening %s!\n", lastIterName)
		return
	}
	defer lastIterfasta.Close()

	var strainNum int
	var rowNum int
	var numSuffix string
	var out []string
	iterReader := bufio.NewReader(lastIterfasta)
	for {
		line, readErr := iterReader.ReadString('\n')
		line = strings.Trim(line, "\r\n")
		line = strings.Trim(line, "\n")
		if len(line) != 0 && !strings.HasPrefix(line, "PG_b") {
			// 每读一行，rowNum就加1，代表行序
			rowNum += 1
			lineBlocks := strings.Split(line, "\t")
			for _, v := range lineBlocks {
				if v == "-" {
					out = append(out, v)
				} else {
					if len(strseqMap[v]) != 0 {
						out = append(out, strseqMap[v])
					} else {
						out = append(out, v)
					}
					strainNum += 1
				}
			}
			if len(strconv.Itoa(rowNum)) == 1 {
				numSuffix = "000" + strconv.Itoa(rowNum)
			} else if len(strconv.Itoa(rowNum)) == 2 {
				numSuffix = "00" + strconv.Itoa(rowNum)
			} else if len(strconv.Itoa(rowNum)) == 3 {
				numSuffix = "0" + strconv.Itoa(rowNum)
			} else if len(strconv.Itoa(rowNum)) >= 4 {
				numSuffix = strconv.Itoa(rowNum)
			}
			fmt.Println(strconv.Itoa(strainNum) + "PG" + numSuffix + "\t" + strings.Join(out, "\t"))
			// _ = strconv.Itoa(strainNum) + "PG" + numSuffix + "\t" + strings.Join(out, "\t")
			numSuffix = "Error"
			out = []string{}
			strainNum = 0
		}
		if readErr == io.EOF {
			break
		}
	}
}

// 输入文件夹路径，输出该路径下所有 子文件 的名称，忽略子文件夹
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


func main() {
    if len(os.Args) > 1 && strings.HasPrefix(os.Args[1], "__internal__") {
        cmd := strings.TrimPrefix(os.Args[1], "__internal__")
        switch cmd {
        case "seqCount":
             seqCountApp(os.Args[1:])
        case "strFamap":
             strFamapApp(os.Args[1:])
        case "batchseqMerge":
             batchseqMergeApp(os.Args[1:])
        case "blastFilter":
             blastFilterApp(os.Args[1:])
        case "PGtmp":
             PGtmpApp(os.Args[1:])
        case "PGtmp2fasta":
             PGtmp2fastaApp(os.Args[1:])
        case "chang_seqid_pgid":
             chang_seqid_pgidApp(os.Args[1:])
        case "iterseqMerge":
             iterseqMergeApp(os.Args[1:])
        case "iterblastFilter":
             iterblastFilterApp(os.Args[1:])
        case "iterPGtmp":
             iterPGtmpApp(os.Args[1:])
        case "iterPGtmp2fasta":
             iterPGtmp2fastaApp(os.Args[1:])
        }
        return
    }
    mainPG(os.Args)
}


func mainPG(args []string) {
	usage := "\nPG.go: Introduction \n\n" + "USAGE:  ./PG SEQ_DIR SIMILARITY Number_of_threads [--yes|-y]\n\n" + "EXAMPLE: $./PG  test  0.7 30\n" + "         $./PG  test  0.7 30 --yes   (non-interactive, auto-delete result folder)\n\n"
	if len(args) < 4 {
		fmt.Printf("%s", usage)
		return
	}
	// seqDir := "result/prot_file_prefix"
	// sim := "0.7"
	// threads := strconv.Itoa(runtime.NumCPU() - 3)
	seqDir := args[1]
	if !strings.HasPrefix(seqDir, "/") && !strings.HasPrefix(seqDir, "./") {
		seqDir = "./" + seqDir
	}
	sim := args[2]
	threads := args[3]

	// Check for --yes / -y flag for non-interactive mode
	forceYes := false
	for _, a := range args[4:] {
		if a == "--yes" || a == "-y" {
			forceYes = true
			break
		}
	}

	seqFiles, err1 := ListDir(seqDir, ".fasta")
	if err1 != nil {
		fmt.Printf("Error happened in the file listing step!\n")
		return
	}
	// 判断结果文件夹是否存在，是否删除结果文件夹
	fmt.Println("*******Make sure to remove the `result` folder before running the BactPG2.0*******")
	_, err := os.Stat("./result")
	if err == nil {
		if forceYes {
			fmt.Println("`result` folder exists, removing automatically (--yes mode).")
			cmd := exec.Command("rm", "-rf", "result")
			err0 := cmd.Run()
			if err0 != nil {
				fmt.Printf("Error: %s\n", err0)
			}
		} else {
			fmt.Println("`result`文件夹存在,是否删除(Y/N)?")
			var Yes_No string
			fmt.Scanln(&Yes_No)
			if Yes_No == "Y" || Yes_No == "y" {
				fmt.Println("删除`result`文件夹。")
				cmd := exec.Command("rm", "-rf", "result")
				err0 := cmd.Run()
				if err0 != nil {
					fmt.Printf("Error: %s\n", err0)
				}
			} else {
				fmt.Println("不删除`result`文件夹，退出。")
				return
			}
		}
	}
	// 创建目录a、b、c，如果目录已经存在则不会重复创建
	mkdirErr := os.MkdirAll("./result/cd_hit_out", 0755)
	if mkdirErr != nil {
		panic(mkdirErr)
	}

	mkdirErr = os.MkdirAll("./result/nr_seq", 0755)
	if mkdirErr != nil {
		panic(mkdirErr)
	}

	mkdirErr = os.MkdirAll("./result/str_famap", 0755)
	if mkdirErr != nil {
		panic(mkdirErr)
	}

	// CD-HIT
	var filePrefix string
	for _, seqFile := range seqFiles {
		if len(seqFile) > 0 {
			filePrefix = strings.Replace(seqFile, ".fasta", "", -1)
			filePrefix = strings.Replace(filePrefix, seqDir, "", -1)
			filePrefix = strings.Replace(filePrefix, "/", "", -1)
			cdhitPrefix := "./result/cd_hit_out/" + filePrefix
			nrseqDir := "./result/nr_seq/" + filePrefix + ".fasta"
			strfamapDir := "./result/str_famap/" + filePrefix + ".map"
			cmd := exec.Command("cd-hit", "-i", seqFile, "-o", cdhitPrefix, "-c", sim, "-T", threads, "-d", "1000", "-aS", sim)
			err2 := cmd.Run()
			if err2 != nil {
				fmt.Printf("Error: %s\n", err2)
			}
			fmt.Printf("CD-HIT for %s was finished\n", filePrefix)

			cmd = exec.Command(args[0], "__internal__strFamap", cdhitPrefix, cdhitPrefix+".clstr")
			buf, err3 := cmd.Output()
			if err3 != nil {
				log.Fatal(err3)
			} else {
				writeFile(strfamapDir, string(buf))
			}

			cmd = exec.Command("cp", cdhitPrefix, nrseqDir)
			err2 = cmd.Run()
			if err2 != nil {
				fmt.Printf("Error: %s\n", err2)
			}
		}
	}

	// superParamer
	var strainseqCount, strCount string
	var cumCount, intCount int
	i := 1
	mkdirErr = os.MkdirAll("./result/batch/batch1", 0755)
	if mkdirErr != nil {
		panic(mkdirErr)
	}
	mkdirErr = os.MkdirAll("./result/iter/iter2", 0755)
	if mkdirErr != nil {
		panic(mkdirErr)
	}

	// // 这里的目的是为了让每个batch内的菌株都是随机的，而不是按照字母顺序排列的
	// // 设置随机种子，以确保每次运行都产生不同的结果
	// rand.Seed(time.Now().UnixNano())
	// // 使用rand.Perm()生成一个0到len(seqFiles)-1的随机排列
	// rand.Shuffle(len(seqFiles), func(i, j int) {
	// 	seqFiles[i], seqFiles[j] = seqFiles[j], seqFiles[i]
	// })

	for j, seqFile := range seqFiles {
		if len(seqFile) > 0 {
			filePrefix = strings.Replace(seqFile, ".fasta", "", -1)
			filePrefix = strings.Replace(filePrefix, seqDir, "", -1)
			filePrefix = strings.Replace(filePrefix, "/", "", -1)
			nrseqDir := "./result/nr_seq/" + filePrefix + ".fasta"
			superparDir := "./result/superParamer.txt"

			cmd := exec.Command(args[0], "__internal__seqCount", nrseqDir)
			buf, err3 := cmd.Output()
			if err3 != nil {
				log.Fatal(err3)
			} else {
				strCount = string(buf)
				intCount, _ = strconv.Atoi(strCount)
				cumCount = cumCount + intCount
				// 这里是batch的阈值，最后应改为100000
				if (cumCount > 50000) && (j != len(seqFiles)-1) {
					i += 1
					cumCount = intCount
					mkdirErr = os.MkdirAll("./result/batch/batch"+strconv.Itoa(i), 0755)
					if mkdirErr != nil {
						panic(mkdirErr)
					}
					if i >= 3 {
						mkdirErr = os.MkdirAll("./result/iter/iter"+strconv.Itoa(i), 0755)
						if mkdirErr != nil {
							panic(mkdirErr)
						}
					}
				} else if j == len(seqFiles)-1 {
					strainseqCount += "batch" + strconv.Itoa(i) + "\t" + filePrefix + "\t" + strCount + "\n"
					writeFile(superparDir, strainseqCount)
					fmt.Printf("superParamer.txt was created!\n")
					break
				}
				strainseqCount += "batch" + strconv.Itoa(i) + "\t" + filePrefix + "\t" + strCount + "\n"
			}
		}
	}

	// batchseqMerge
	batchMap := make(map[string]string)
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
		batchMap[lineBlocks[1]] = lineBlocks[0]
	}

	var batchNum0, batchNum, seqs string
	var j int
	for k, seqFile := range seqFiles {
		if len(seqFile) > 0 {
			filePrefix = strings.Replace(seqFile, ".fasta", "", -1)
			filePrefix = strings.Replace(filePrefix, seqDir, "", -1)
			filePrefix = strings.Replace(filePrefix, "/", "", -1)
			nrseqDir := "./result/nr_seq/" + filePrefix + ".fasta"
			batchNum0 = batchMap[filePrefix]
			batchseqDir0 := "./result/batch/" + batchNum0 + "/" + batchNum0 + "_nr.fasta"
			batchseqDir := "./result/batch/" + batchNum + "/" + batchNum + "_nr.fasta"

			if batchNum0 != batchNum {
				if len(seqs) != 0 {
					writeFile(batchseqDir, seqs)
					fmt.Printf("Seqs of %s was merged!\n", batchNum)
					seqs = ""
				}
				batchNum = batchNum0
				j = 1
				cmd := exec.Command(args[0], "__internal__batchseqMerge", nrseqDir, strconv.Itoa(j))
				buf, err3 := cmd.Output()
				if err3 != nil {
					log.Fatal(err3)
				} else {
					seqs = string(buf)
				}
			} else {
				j += 1
				cmd := exec.Command(args[0], "__internal__batchseqMerge", nrseqDir, strconv.Itoa(j))
				buf, err3 := cmd.Output()
				if err3 != nil {
					log.Fatal(err3)
				} else {
					seqs += string(buf)
					// 循环到了最后一个batch的最后一个文件，直接输出
					if k == len(seqFiles)-1 {
						writeFile(batchseqDir0, seqs)
						fmt.Printf("Seqs of %s was merged!\n", batchNum0)
						seqs = ""
					}
				}
			}
		}
	}

	// blast
	batchsDir := "./result/batch"
	subFolders, batchErr := ListSubfile(batchsDir)
	if batchErr != nil {
		fmt.Printf("Error happened in the batchfile listing step!\n")
		return
	}
	subFolders = sortSlice(subFolders)
	for _, subFolder := range subFolders {
		db_name := batchsDir + "/" + subFolder + "/" + "blast_out" + "/" + subFolder + "_nr"
		nrFaa := batchsDir + "/" + subFolder + "/" + subFolder + "_nr.fasta"
		blastoutFile := batchsDir + "/" + subFolder + "/" + subFolder + "_blastout.txt"
		// blast database building ...
		cmd := exec.Command("makeblastdb", "-in", nrFaa, "-input_type", "fasta", "-dbtype", "prot", "-out", db_name)
		err := cmd.Run()
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}
		fmt.Printf("Blastdb of %s was created!\n", subFolder)

		// making alignment
		cmd = exec.Command("blastp", "-query", nrFaa, "-db", db_name, "-evalue", "1e-5", "-outfmt", "6", "-out", blastoutFile, "-num_threads", threads)
		fmt.Printf("Waiting for blast command to finish for %s...\n", subFolder)
		err = cmd.Run()
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}
		fmt.Printf("Blast for %s have finished!\n", subFolder)
	}

	// blast_out_filter & pg_tmp_file
	for _, subFolder := range subFolders {
		blastoutFile := batchsDir + "/" + subFolder + "/" + subFolder + "_blastout.txt"
		blastoutFilter := batchsDir + "/" + subFolder + "/" + subFolder + "_blastoutFilter.txt"
		PGtmp1Dir := batchsDir + "/" + subFolder + "/" + subFolder + ".PG.tmp1.txt"
		batchPGtmpfasta := batchsDir + "/" + subFolder + "/" + subFolder + ".PG.tmp.fasta"

		cmd := exec.Command(args[0], "__internal__blastFilter", blastoutFile, sim)
		buf, err := cmd.Output()
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		} else {
			writeFile(blastoutFilter, string(buf))
			fmt.Printf("Best hits were picked from %s!\n", subFolder)
		}

		cmd = exec.Command(args[0], "__internal__PGtmp", blastoutFilter)
		buf, err = cmd.Output()
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		} else {
			writeFile(PGtmp1Dir, string(buf))
			fmt.Printf("PG.tmp1 file was created for %s!\n", subFolder)
		}

		cmd = exec.Command(args[0], "__internal__PGtmp2fasta", PGtmp1Dir)
		buf, err = cmd.Output()
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		} else {
			writeFile(batchPGtmpfasta, string(buf))
			fmt.Printf("PG.tmp.fasta was created for %s!\n", subFolder)
		}
	}

	// batchs seq merge to iter & blast_out_filter & pg_tmp_file
	if len(subFolders) == 1 {
		// 如果只有一个batch，就不需要做iterseqMerge了
		fmt.Println("There is only one batch, no need to do iterseqMerge!")
		PGtxt := "./result/PG.txt"
		// 将batch1.PG.tmp1.txt改名为PG.txt
		cmd := exec.Command(args[0], "__internal__chang_seqid_pgid", batchsDir+"/batch1/batch1.PG.tmp1.txt")
		buf, err := cmd.Output()
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		} else {
			writeFile(PGtxt, string(buf))
			fmt.Println("Final PG was created!")
		}
		return
	}
	for k, subFolder := range subFolders {
		k = k + 1
		batchPGtmpfasta := batchsDir + "/" + subFolder + "/" + subFolder + ".PG.tmp.fasta"
		iterPGtmpfasta0 := "./result/iter/iter" + strconv.Itoa(k-1) + "/iter" + strconv.Itoa(k-1) + ".PG.tmp.fasta"
		iterPGtmpfasta := "./result/iter/iter" + strconv.Itoa(k) + "/iter" + strconv.Itoa(k) + ".PG.tmp.fasta"
		iterfasta := "./result/iter/iter" + strconv.Itoa(k) + "/iter" + strconv.Itoa(k) + "_nr.fasta"
		db_name := "./result/iter/iter" + strconv.Itoa(k) + "/blast_out" + "/iter" + strconv.Itoa(k)
		blastoutFile := "./result/iter/iter" + strconv.Itoa(k) + "/iter" + strconv.Itoa(k) + "_blastout.txt"
		blastoutFilter := "./result/iter/iter" + strconv.Itoa(k) + "/iter" + strconv.Itoa(k) + "_blastoutFilter.txt"
		PGtmp1Dir := "./result/iter/iter" + strconv.Itoa(k) + "/iter" + strconv.Itoa(k) + ".PG.tmp1.txt"

		if k == 1 {
			cmd := exec.Command(args[0], "__internal__iterseqMerge", batchPGtmpfasta)
			buf, err := cmd.Output()
			if err != nil {
				log.Fatal(err)
			} else {
				seqs = string(buf)
				continue
			}
		} else if k == 2 {
			cmd := exec.Command(args[0], "__internal__iterseqMerge", batchPGtmpfasta)
			buf, err := cmd.Output()
			if err != nil {
				log.Fatal(err)
			} else {
				seqs += string(buf)
				writeFile(iterfasta, seqs)
				fmt.Printf("iter%s_nr.fasta was created!\n", strconv.Itoa(k))
				seqs = ""
			}
		} else if k >= 3 {
			// 把上一个iter去冗余序列打印出来
			cmd := exec.Command(args[0], "__internal__iterseqMerge", iterPGtmpfasta0)
			buf, err := cmd.Output()
			if err != nil {
				log.Fatal(err)
			} else {
				seqs = string(buf)
			}
			// 把现在这个batch去冗余序列打印出来
			cmd = exec.Command(args[0], "__internal__iterseqMerge", batchPGtmpfasta)
			buf, err = cmd.Output()
			if err != nil {
				log.Fatal(err)
			} else {
				seqs += string(buf)
				writeFile(iterfasta, seqs)
				fmt.Printf("iter%s_nr.fasta was created!\n", strconv.Itoa(k))
				seqs = ""
			}
		}

		// blast database building ...
		cmd := exec.Command("makeblastdb", "-in", iterfasta, "-input_type", "fasta", "-dbtype", "prot", "-out", db_name)
		err := cmd.Run()
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}
		fmt.Printf("Blastdb of iter%s was created!\n", strconv.Itoa(k))

		// making alignment
		cmd = exec.Command("blastp", "-query", iterfasta, "-db", db_name, "-evalue", "1e-5", "-outfmt", "6", "-out", blastoutFile, "-num_threads", threads)
		fmt.Printf("Waiting for blast command to finish for iter%s...\n", strconv.Itoa(k))
		err = cmd.Run()
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}
		fmt.Printf("Blast for iter%s have finished!\n", strconv.Itoa(k))

		// blast_out_filter
		cmd = exec.Command(args[0], "__internal__iterblastFilter", blastoutFile, sim)
		buf, err := cmd.Output()
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		} else {
			writeFile(blastoutFilter, string(buf))
			fmt.Printf("Best hits were picked from iter%s!\n", strconv.Itoa(k))
		}

		// PGtmp1 of iterk
		cmd = exec.Command(args[0], "__internal__iterPGtmp", blastoutFilter)
		buf, err = cmd.Output()
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		} else {
			writeFile(PGtmp1Dir, string(buf))
			fmt.Printf("PG.tmp1 file was created for iter%s!\n", strconv.Itoa(k))
		}

		// PGtmpfasta of iterk
		cmd = exec.Command(args[0], "__internal__iterPGtmp2fasta", PGtmp1Dir)
		buf, err = cmd.Output()
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		} else {
			writeFile(iterPGtmpfasta, string(buf))
			fmt.Printf("PG.tmp.fasta was created for iter%s!\n", strconv.Itoa(k))
		}
	}

	// Filnal PG
	itersDir := "./result/iter"
	iters, iterErr := ListSubfile(itersDir)
	if iterErr != nil {
		fmt.Printf("Error happened in the iterfile listing step!\n")
		return
	}
	// 按照数字排序
	iters = sortSlice(iters)
	// 取最后一个 iter.PG.tmp1.txt 作为 finalPG
	lastIterName := iters[len(iters)-1]
	lastIterfile := "./result/iter/" + lastIterName + "/" + lastIterName + ".PG.tmp1.txt"
	// finalPG := "./result/finalPG.txt"
	PGtxt := "./result/PG.txt"

	cmd := exec.Command(args[0], "__internal__chang_seqid_pgid", lastIterfile)
	buf, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	} else {
		// writeFile(finalPG, string(buf))
		writeFile(PGtxt, string(buf))
		fmt.Println("Final PG was created!")
	}
}

// 输入文件夹路径及该路径下文件的后缀，输出该路径下所有文件的完整路径（绝对/相对）
func ListDir(dirPth string, suffix string) (files []string, err error) {
	files = []string{}
	dir, err := os.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}
	PthSep := string(os.PathSeparator)
	suffix = strings.ToUpper(suffix)

	for _, fi := range dir {
		if fi.IsDir() {
			continue
		}
		if strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) {
			files = append(files, dirPth+PthSep+fi.Name())
		}
	}
	return files, nil
}

// 输入文件夹路径，输出该路径下所有 子文件夹 的名称，忽略子文件
func ListSubfile(dirPth string) (files []string, err error) {
	files = []string{}
	dir, err := os.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}

	for _, fi := range dir {
		if fi.IsDir() {
			files = append(files, fi.Name())
		}
	}
	return files, nil
}

func writeFile(newFile string, seq string) {
	outFile, err := os.Create(newFile)
	defer outFile.Close()
	if err != nil {
		fmt.Printf("Errors happened in creating result files!\n")
	}
	outFile.WriteString(seq)
}

func sortSlice(strSlice []string) []string {
	sort.Slice(strSlice, func(i, j int) bool {
		if strings.Contains(strSlice[i], "batch") {
			b1 := strings.TrimLeft(strSlice[i], "batch")
			b2 := strings.TrimLeft(strSlice[j], "batch")
			if b1 != strSlice[i] && b2 != strSlice[j] {
				n1, err1 := strconv.Atoi(b1)
				n2, err2 := strconv.Atoi(b2)
				if err1 == nil && err2 == nil {
					if n1 < n2 {
						return true
					} else if n1 > n2 {
						return false
					}
				}
			}
		} else if strings.Contains(strSlice[i], "iter") {
			b1 := strings.TrimLeft(strSlice[i], "iter")
			b2 := strings.TrimLeft(strSlice[j], "iter")
			if b1 != strSlice[i] && b2 != strSlice[j] {
				n1, err1 := strconv.Atoi(b1)
				n2, err2 := strconv.Atoi(b2)
				if err1 == nil && err2 == nil {
					if n1 < n2 {
						return true
					} else if n1 > n2 {
						return false
					}
				}
			}
		}
		return strSlice[i] < strSlice[j]
	})
	return strSlice
}
