/*
 * @Usage: Usage of this script
 */
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
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
