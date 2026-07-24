//simcovfilter
//USAGE: ./simcovfilter DB_PROT_LEN BEST_HIT SIM COVERAGE >BLAST_FILTER

package simcov

import (
  "fmt"
  "os"
  "io"
  "bufio"
  "strings"
  "strconv"
)

func RunSimCov () {
  usage := "simcovfilter DB_PROT_LEN BEST_HIT SIM COVERAGE"
  if len(os.Args) < 5 {
    fmt.Printf("Argument not right!\n")
    fmt.Printf("USAGE: %s\n",usage)
    return
  }

  dbProtLenName := os.Args[1]
  bestHitName := os.Args[2]
  simCutoff,_ := strconv.ParseFloat(os.Args[3],32)
//  simCutoff = float32(simCutoff0/100)
  covCutoff,_ := strconv.ParseFloat(os.Args[4],32)
//  covCutoff = float32(covCutoff)

  dbProtLenFile, openErr1 := os.Open(dbProtLenName)
  if openErr1 != nil {
    fmt.Printf("Error happened in opening %s!\n",dbProtLenName)
    return
  }
  defer dbProtLenFile.Close()

  dbProtLenReader := bufio.NewReader(dbProtLenFile)
  dbProtLen := make(map[string]int)
  for {
    line, readErr1 := dbProtLenReader.ReadString('\n')
    line = strings.Trim(line,"\r\n")
    line = strings.Trim(line,"\n")
    if len(line) != 0 {
      lineArray := strings.Split(line,"\t")
      dbProtLen[lineArray[0]],_ = strconv.Atoi(lineArray[1])
    }  
    if readErr1 == io.EOF {
      break
    }
  }

  bestHitFile, openErr4 := os.Open(bestHitName)
  if openErr4 != nil {
    fmt.Printf("Error happened in opening %s!\n",bestHitName)
    return
  }
  defer bestHitFile.Close()

  bestHitReader := bufio.NewReader(bestHitFile)
  fmt.Printf("Query\tSubject\tCoverage\tSimilarity\n")
  for {
    line, readErr4 := bestHitReader.ReadString('\n')
    line = strings.Trim(line,"\r\n")
    line = strings.Trim(line,"\n")
    if len(line) != 0 {
      lineArray4 := strings.Split(line,"\t")
      sim,_ := strconv.ParseFloat(lineArray4[2],32)
      sim = sim/100
      dbSt,_ := strconv.Atoi(lineArray4[8])
      dbEn,_ := strconv.Atoi(lineArray4[9])
      covLen := dbEn - dbSt + 1
      lenCov := float64(covLen) / float64(dbProtLen[lineArray4[1]])
      if lenCov >= covCutoff && sim >= simCutoff {
        fmt.Printf("%s\t%s\t%.2f\t%.2f\n",lineArray4[0],lineArray4[1],lenCov,sim)
      }
    }  
    if readErr4 == io.EOF {
      break
    }
  } 	
}