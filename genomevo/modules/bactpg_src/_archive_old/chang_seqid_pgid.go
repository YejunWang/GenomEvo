package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

func chang_seqid_pgidApp(args []string) {
	if len(args) < 2 {
		fmt.Printf("No input file!\n")
		return
	}

	lastIterName := args[1]

	// 读取所有菌株的同菌株同源家族蛋白名，生成map
	strseqMap := make(map[string]string)
	str_famap := "./result/str_famap"
	str_famapList, str_famapErr := ioutil.ReadDir(str_famap)
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
func ListSubfile(dirPth string) (files []string, err error) {
	files = []string{}
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}

	for _, fi := range dir {
		if !fi.IsDir() {
			files = append(files, fi.Name())
		}
	}
	return files, nil
}
