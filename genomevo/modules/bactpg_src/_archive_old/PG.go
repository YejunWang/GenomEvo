package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

func main() {
	usage := "\nPG.go: Introduction \n\n" + "USAGE:  ./PG SEQ_DIR SIMILARITY Number_of_threads\n\n" + "EXAMPLE: $./PG  test  0.7 30\n\n"
	if len(os.Args) < 4 {
		fmt.Printf("%s", usage)
		return
	}
	// seqDir := "result/prot_file_prefix"
	// sim := "0.7"
	// threads := strconv.Itoa(runtime.NumCPU() - 3)
	seqDir := "./" + os.Args[1]
	sim := os.Args[2]
	threads := os.Args[3]
	seqFiles, err1 := ListDir(seqDir, ".fasta")
	if err1 != nil {
		fmt.Printf("Error happened in the file listing step!\n")
		return
	}
	// 判断结果文件夹是否存在，是否删除结果文件夹
	fmt.Println("*******Make sure to remove the `result` folder before running the BactPG2.0*******")
	var Yes_No string
	_, err := os.Stat("./result")
	if err == nil {
		fmt.Println("`result`文件夹存在,是否删除(Y/N)?")
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

			cmd = exec.Command("./bin/strFamap", cdhitPrefix, cdhitPrefix+".clstr")
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

			cmd := exec.Command("./bin/seqCount", nrseqDir)
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
				cmd := exec.Command("./bin/batchseqMerge", nrseqDir, strconv.Itoa(j))
				buf, err3 := cmd.Output()
				if err3 != nil {
					log.Fatal(err3)
				} else {
					seqs = string(buf)
				}
			} else {
				j += 1
				cmd := exec.Command("./bin/batchseqMerge", nrseqDir, strconv.Itoa(j))
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

		cmd := exec.Command("./bin/blastFilter", blastoutFile, sim)
		buf, err := cmd.Output()
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		} else {
			writeFile(blastoutFilter, string(buf))
			fmt.Printf("Best hits were picked from %s!\n", subFolder)
		}

		cmd = exec.Command("./bin/PGtmp", blastoutFilter)
		buf, err = cmd.Output()
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		} else {
			writeFile(PGtmp1Dir, string(buf))
			fmt.Printf("PG.tmp1 file was created for %s!\n", subFolder)
		}

		cmd = exec.Command("./bin/PGtmp2fasta", PGtmp1Dir)
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
		cmd := exec.Command("./bin/chang_seqid_pgid", batchsDir+"/batch1/batch1.PG.tmp1.txt")
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
			cmd := exec.Command("./bin/iterseqMerge", batchPGtmpfasta)
			buf, err := cmd.Output()
			if err != nil {
				log.Fatal(err)
			} else {
				seqs = string(buf)
				continue
			}
		} else if k == 2 {
			cmd := exec.Command("./bin/iterseqMerge", batchPGtmpfasta)
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
			cmd := exec.Command("./bin/iterseqMerge", iterPGtmpfasta0)
			buf, err := cmd.Output()
			if err != nil {
				log.Fatal(err)
			} else {
				seqs = string(buf)
			}
			// 把现在这个batch去冗余序列打印出来
			cmd = exec.Command("./bin/iterseqMerge", batchPGtmpfasta)
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
		cmd = exec.Command("./bin/iterblastFilter", blastoutFile, sim)
		buf, err := cmd.Output()
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		} else {
			writeFile(blastoutFilter, string(buf))
			fmt.Printf("Best hits were picked from iter%s!\n", strconv.Itoa(k))
		}

		// PGtmp1 of iterk
		cmd = exec.Command("./bin/iterPGtmp", blastoutFilter)
		buf, err = cmd.Output()
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		} else {
			writeFile(PGtmp1Dir, string(buf))
			fmt.Printf("PG.tmp1 file was created for iter%s!\n", strconv.Itoa(k))
		}

		// PGtmpfasta of iterk
		cmd = exec.Command("./bin/iterPGtmp2fasta", PGtmp1Dir)
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

	cmd := exec.Command("./bin/chang_seqid_pgid", lastIterfile)
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
	dir, err := ioutil.ReadDir(dirPth)
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
	dir, err := ioutil.ReadDir(dirPth)
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
