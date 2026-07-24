//CG.go:  to identify the core gene set from the proteome sequence files of multiple bacterial strains
//USAGE:	./CG	SEQ_DIR REF	SIMILARITY COVERAGE

/* The pre－compiled 'CG' program should be put in current directory together with 'Tool/BactCG2.0/bin' folder where pre-compiled 'lenExt',
   'bestpicker', 'simcovfilter', 'mutbest' and 'combSingleGeneOrtho' programs. Standalone BLAST (>= version 2.2.30+) should
   be installed and set in environmental variables.

   Put your FASTA format sequence files in a subfolder secondary to the current directory. The proteome sequences of each
   strain are curated in a single FASTA file named as 'xxx.fasta', where 'xxx' is suggested to be replaced by the simplified
   strain name.

   To identify the core genome, one reference proteome file is required to be set as the reference proteome, and its strain
   name exactly represented in the sequence file name (before '.fasta') is set as the reference strain name. The cutoffs for
   sequence similarity and the ratio of coverage aligned-protein length against the whole aligned-protein length are set by
   the users directly as the argument.

   For example, given sequence files stored in a sub-folder named as 'seq', and the cutoff for similarity and coverage being
   set as 0.8 and 0.9 respectively, the following command should be input after moving into the current working directory:

   $ ./CG  seq  LT2  0.8  0.9

   The program will generate a new sub-folder 'result' in current directory, and the final core gene set file could be found
   in the directory './result/cg_result/' and is named as 'CG.tab.txt'.
*/

package cg

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

func RunCG() {
	threads := strconv.Itoa(runtime.NumCPU() - 1)
	usage := "\nCG.go: Identifying core genome from multiple proteome sequence files. \n\n" + "USAGE:  ./CG SEQ_DIR REF SIMILARITY COVERAGE\n\n" + "EXAMPLE: $./CG  test  LT2  0.8  0.9\n\n"
	if len(os.Args) < 5 {
		fmt.Printf("%s", usage)
		return
	}
	seqDir := os.Args[1]
	ref := os.Args[2]
	sim := os.Args[3]
	cov := os.Args[4]
	seqFiles, err1 := ListDir(seqDir, ".fasta")
	if err1 != nil {
		fmt.Printf("Error happened in the file listing step!\n")
		return
	}

	cmd := exec.Command("mkdir", "-p", "result")
	err0 := cmd.Run()
	if err0 != nil {
		fmt.Printf("Error: %s\n", err0)
	}

	cmd = exec.Command("mkdir", "-p", "result/db")
	err0 = cmd.Run()
	if err0 != nil {
		fmt.Printf("Error: %s\n", err0)
	}

	cmd = exec.Command("mkdir", "-p", "result/blast_out")
	err0 = cmd.Run()
	if err0 != nil {
		fmt.Printf("Error: %s\n", err0)
	}

	cmd = exec.Command("mkdir", "-p", "result/prot_len")
	err0 = cmd.Run()
	if err0 != nil {
		fmt.Printf("Error: %s\n", err0)
	}

	cmd = exec.Command("mkdir", "-p", "result/out_best")
	err0 = cmd.Run()
	if err0 != nil {
		fmt.Printf("Error: %s\n", err0)
	}

	cmd = exec.Command("mkdir", "-p", "result/out_best_filt")
	err0 = cmd.Run()
	if err0 != nil {
		fmt.Printf("Error: %s\n", err0)
	}

	cmd = exec.Command("mkdir", "-p", "result/out_mutbest_filt")
	err0 = cmd.Run()
	if err0 != nil {
		fmt.Printf("Error: %s\n", err0)
	}

	cmd = exec.Command("mkdir", "-p", "result/cg_result")
	err0 = cmd.Run()
	if err0 != nil {
		fmt.Printf("Error: %s\n", err0)
	}

	//Preparation of protein_length and blast databases
	var filePrefix, db_name string
	for _, seqFile := range seqFiles {
		if len(seqFile) > 0 {
			//protein_length_db preparation
			filePrefix = strings.Replace(seqFile, ".fasta", "", -1)
			filePrefix = strings.Replace(filePrefix, seqDir, "", -1)
			filePrefix = strings.Replace(filePrefix, "/", "", -1)
			lengthFile := "./result/prot_len/" + filePrefix + ".prot_len.txt"
			cmd = exec.Command(os.Args[0], "lenext", seqFile)
			buf0, err2 := cmd.Output()
			if err2 != nil {
				log.Fatal(err2)
			} else {
				writeFile(lengthFile, string(buf0))
			}
			fmt.Printf("protein length db for %s was created\n", filePrefix)

			//blast database building ...
			db_name = "./result/db/" + filePrefix
			cmd = exec.Command("makeblastdb", "-in", seqFile, "-input_type", "fasta", "-dbtype", "prot", "-out", db_name)
			err2 = cmd.Run()
			if err2 != nil {
				fmt.Printf("Error: %s\n", err2)
			}
			fmt.Printf("Waiting for makeblastdb command to finish for %s...\n", filePrefix)
		}
	}

	refseq := seqDir + "/" + ref + ".fasta"
	refdb := "./result/db/" + ref
	var q_blastout, s_blastout, q_best, s_best, q_best_filt, s_best_filt, out_mutbest_filt, strainNames, out_mutbest_filt0, cgFile0, cgFile string
	seqOrder := 1
	for _, seqFile := range seqFiles {
		if len(seqFile) > 0 {
			filePrefix = strings.Replace(seqFile, ".fasta", "", -1)
			filePrefix = strings.Replace(filePrefix, seqDir, "", -1)
			filePrefix = strings.Replace(filePrefix, "/", "", -1)
			q_blastout = "./result/blast_out/" + filePrefix + ".vs." + ref + ".blastout.m6.txt"
			db_name = "./result/db/" + filePrefix
			q_best = "./result/out_best/" + filePrefix + ".vs." + ref + ".bestHit.txt"
			s_blastout = "./result/blast_out/" + ref + ".vs." + filePrefix + ".blastout.m6.txt"
			s_best = "./result/out_best/" + ref + ".vs." + filePrefix + ".bestHit.txt"
			lengthFile1 := "./result/prot_len/" + filePrefix + ".prot_len.txt"
			lengthFile2 := "./result/prot_len/" + ref + ".prot_len.txt"
			q_best_filt = "./result/out_best_filt/" + filePrefix + ".vs." + ref + ".best.filt.txt"
			s_best_filt = "./result/out_best_filt/" + ref + ".vs." + filePrefix + ".best.filt.txt"
			out_mutbest_filt = "./result/out_mutbest_filt/" + ref + ".vs." + filePrefix + ".mutbest.filt.txt"

			if !strings.EqualFold(filePrefix, ref) {
				//making alignment between target sequence against reference ...
				cmd = exec.Command("blastp", "-query", seqFile, "-db", refdb, "-evalue", "1e-5", "-outfmt", "6", "-out", q_blastout, "-num_threads", threads)
				fmt.Printf("Waiting for blast command to finish for %s against %s...\n", filePrefix, ref)
				err3 := cmd.Run()
				if err3 != nil {
					fmt.Printf("Error: %s\n", err3)
				}

				cmd = exec.Command("blastp", "-query", refseq, "-db", db_name, "-evalue", "1e-5", "-outfmt", "6", "-out", s_blastout, "-num_threads", threads)
				fmt.Printf("Waiting for blast command to finish for %s against %s...\n", ref, filePrefix)
				err4 := cmd.Run()
				if err4 != nil {
					fmt.Printf("Error: %s\n", err4)
				}

				//picking the best hits
				cmd1 := exec.Command(os.Args[0], "bestpicker", q_blastout)
				buf, err1 := cmd1.Output()
				if err1 != nil {
					log.Fatal(err1)
				} else {
					writeFile(q_best, string(buf))
				}
				fmt.Printf("best hits were picked from %s\n", q_blastout)

				cmd2 := exec.Command(os.Args[0], "bestpicker", s_blastout)
				buf, err1 = cmd2.Output()
				if err1 != nil {
					log.Fatal(err1)
				} else {
					writeFile(s_best, string(buf))
				}
				fmt.Printf("best hits were picked from %s\n", s_blastout)

				//filtering best hits with coverage length and identities
				cmd1 = exec.Command(os.Args[0], "simcov", lengthFile2, q_best, sim, cov)
				buf, err1 = cmd1.Output()
				if err1 != nil {
					log.Fatal(err1)
				} else {
					writeFile(q_best_filt, string(buf))
				}
				fmt.Printf("best hits were filtered from %s\n", q_best)
				
				cmd2 = exec.Command(os.Args[0], "simcov", lengthFile1, s_best, sim, cov)
				buf, err1 = cmd2.Output()
				if err1 != nil {
					log.Fatal(err1)
				} else {
					writeFile(s_best_filt, string(buf))
				}
				fmt.Printf("best hits were filtered from %s\n", s_best)

				//Retrieving mutual best hits
				cmd := exec.Command(os.Args[0], "mutbest", s_best_filt, q_best_filt)
				buf, err1 = cmd.Output()
				if err1 != nil {
					log.Fatal(err1)
				} else {
					writeFile(out_mutbest_filt, string(buf))
				}
				fmt.Printf("mutual best hits were retrieved for %s\n", filePrefix)

				//CG extracting
				if seqOrder == 1 {
					out_mutbest_filt0 = out_mutbest_filt
					strainNames = ref + "\t" + filePrefix
				} else if seqOrder == 2 {
					strainNames += "\t" + filePrefix
					cgFile = strconv.Itoa(seqOrder)
					cgFile += ".cg.txt"
					cmd = exec.Command(os.Args[0], "combsingle", out_mutbest_filt0, out_mutbest_filt, ">", cgFile)
					buf, err1 = cmd.Output()
					if err1 != nil {
						log.Fatal(err1)
					} else {
						writeFile(cgFile, string(buf))
					}
					fmt.Printf("CG was retrieved for %s\n", filePrefix)
					cgFile0 = cgFile
				} else {
					strainNames += "\t" + filePrefix
					cgFile = strconv.Itoa(seqOrder)
					cgFile += ".cg.txt"
					cgFile = "./result/cg_result/" + cgFile
					cmd = exec.Command(os.Args[0], "combsingle", cgFile0, out_mutbest_filt)
					buf, err1 = cmd.Output()
					if err1 != nil {
						log.Fatal(err1)
					} else {
						writeFile(cgFile, string(buf))
					}
					fmt.Printf("CG was retrieved for %s\n", filePrefix)
					cmd = exec.Command("rm", cgFile0)
					err1 = cmd.Run()
					if err1 != nil {
						fmt.Printf("Error happened in removing the temporary %s: %s\n", cgFile0, err1)
					}
					cgFile0 = cgFile
				}
				seqOrder++
			}
		}
	}
	cgFileOpen, openErr1 := os.Open(cgFile)
	if openErr1 != nil {
		fmt.Printf("Error happened in opening %s: %s\n", cgFile, openErr1)
	}
	defer cgFileOpen.Close()

	cgReader := bufio.NewReader(cgFileOpen)
	cgLines := strainNames + "\n"
	cgResult := "./result/cg_result/" + "CG.tab.txt"

	for {
		cgLine0, readErr1 := cgReader.ReadString('\n')
		cgLines += cgLine0
		if readErr1 == io.EOF {
			writeFile(cgResult, cgLines)
			break
		}
	}
	cmd = exec.Command("rm", cgFile)
	err1 = cmd.Run()
	if err1 != nil {
		fmt.Printf("Error happened in removing the temporary %s: %s\n", cgFile, err1)
	}
	fmt.Printf("CG tab file was exported as %s\n", cgResult)
}

func ListDir(dirPth string, suffix string) (files []string, err error) {
	files = make([]string, 0, 1000)
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

func writeFile(newFile string, seq string) {
	outFile, err := os.Create(newFile)
	defer outFile.Close()
	if err != nil {
		fmt.Printf("Errors happened in creating result files!\n")
	}
	outFile.WriteString(seq)
}
