package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func runParse(gbFileName string) {
	gbFile, err := os.Open(gbFileName)
	if err != nil {
		fmt.Printf("Error happened in opening %s!\n", gbFileName)
		return
	}
	defer gbFile.Close()

	scanner := bufio.NewScanner(gbFile)

	var seq, molType, geneAcc, gene, locusTag, product, geneID string
	var flag, flag0, flag2 int

	for scanner.Scan() {
		line := scanner.Text()
		
		pat01, _ := regexp.MatchString("  rRNA  ", line)
		pat02, _ := regexp.MatchString("  tRNA  ", line)
		pat03, _ := regexp.MatchString("  misc_feature  ", line)
		pat04, _ := regexp.MatchString("  ncRNA  ", line)
		patA, _ := regexp.MatchString("/gene=", line)
		patB, _ := regexp.MatchString("/locus_tag=", line)
		patC, _ := regexp.MatchString("/product=", line)
		patD, _ := regexp.MatchString("/EC_number=", line)
		patF, _ := regexp.MatchString("                     /pseudo", line)
		
		pat1, _ := regexp.MatchString("  CDS  ", line)
		patE, _ := regexp.MatchString("/translation=", line)

		if pat01 {
			molType = "rRNA"
			line = strings.Replace(line, "rRNA", "", 1)
			geneAcc = seqTrim(line)
			flag0 = 1
			gene, geneID, locusTag, product = "", "", "", ""
		} else if pat02 {
			molType = "tRNA"
			line = strings.Replace(line, "tRNA", "", 1)
			geneAcc = seqTrim(line)
			flag0 = 1
			gene, geneID, locusTag, product = "", "", "", ""
		} else if pat03 {
			molType = "misc_feature"
			line = strings.Replace(line, "misc_feature", "", 1)
			geneAcc = seqTrim(line)
			flag0 = 1
			gene, geneID, locusTag, product = "", "", "", ""
		} else if pat04 {
			molType = "ncRNA"
			line = strings.Replace(line, "ncRNA", "", 1)
			geneAcc = seqTrim(line)
			flag0 = 1
			gene, geneID, locusTag, product = "", "", "", ""
		} else if flag0 == 1 {
			if patA {
				gene = strings.Replace(line, "/gene=\"", "", -1)
				gene = strings.Replace(gene, "\"", "", -1)
				gene = seqTrim(gene)
			}
			if patB {
				locusTag = strings.Replace(line, "/locus_tag=\"", "", -1)
				locusTag = strings.Replace(locusTag, "\"", "", -1)
				locusTag = seqTrim(locusTag)
			}
			if patC {
				product = strings.Replace(line, "                     /product=\"", "", -1)
				product = strings.Replace(product, "\"", "", -1)
				// product = strings.Replace(product, "\n", "", -1)
				flag0 = 0
				fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\t\n", geneAcc, molType, locusTag, geneID, gene, product)
				gene, geneID, locusTag, product = "", "", "", ""
			}
		}

		if pat1 {
			molType = "CDS"
			line = strings.Replace(line, "CDS", "", 1)
			geneAcc = seqTrim(line)
			flag0 = 0
			flag = 1
			gene, geneID, locusTag, product = "", "", "", ""
		} else if flag == 1 {
			if patA {
				gene = strings.Replace(line, "/gene=\"", "", -1)
				gene = strings.Replace(gene, "\"", "", -1)
				gene = seqTrim(gene)
			}
			if patB {
				locusTag = strings.Replace(line, "/locus_tag=\"", "", -1)
				locusTag = strings.Replace(locusTag, "\"", "", -1)
				locusTag = seqTrim(locusTag)
			}
			if patC {
				product = strings.Replace(line, "                     /product=\"", "", -1)
				product = strings.Replace(product, "\"", "", -1)
			}
			if patD {
				geneID = strings.Replace(line, "/EC_number=\"", "", -1)
				geneID = strings.Replace(geneID, "\"", "", -1)
				geneID = seqTrim(geneID)
			}
			if patE {
				line = strings.Replace(line, "/translation=\"", "", -1)
				line = seqTrim(line)
				flag = 0
				if strings.Contains(line, "\"") {
					seq = strings.Replace(line, "\"", "", 1)
					fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\t%s\n", geneAcc, molType, locusTag, geneID, gene, product, seq)
					gene, geneID, locusTag, product, seq = "", "", "", "", ""
				} else {
					flag2 = 1
					seq = line
				}
			}
			if patF {
				fmt.Printf("%s\t%s\t%s\t%s\t%s\tpseudo\n", geneAcc, molType, locusTag, geneID, gene)
				gene, geneID, locusTag, product, seq = "", "", "", "", ""
				flag = 0
			}
		} else if flag2 == 1 {
			line = seqTrim(line)
			if strings.Contains(line, "\"") {
				line = strings.Replace(line, "\"", "", -1)
				seq += line
				fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\t%s\n", geneAcc, molType, locusTag, geneID, gene, product, seq)
				flag2 = 0
				gene, geneID, locusTag, product, seq = "", "", "", "", ""
			} else {
				seq += line
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning file: %v\n", err)
	}
}

func seqTrim(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Replace(s, " ", "", -1)
	return s
}
