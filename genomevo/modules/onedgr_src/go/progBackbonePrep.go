//progBackbonePrep.go: to modify the backbone file of progressive Mauve before further analysis
//USAGE: ./progBackbonePrep  PROG_MAUVE_BACKBONE_FILE  >MAUVE_BACKBONE_FILE


// seq0_leftend seq0_rightend seq1_leftend  seq1_rightend
// 363106 364287  0 0

package main

import (
  "fmt"
  "io"
  "bufio"
  "strings"
  "strconv"
  "os"
)

func main () {
  usage := "progBackbonePrep: to modify the backbone file of progressive Mauve before further analysis.\n\n" + "./progBackbonePrep  PROG_MAUVE_BACKBONE_FILE  >MAUVE_BACKBONE_FILE\n\n"
  if len(os.Args) < 2 {
    fmt.Printf("%s",usage)
    return
  }
  progName := os.Args[1]

  prog, openErr := os.Open(progName)
  if openErr != nil {
    fmt.Printf("Error happened in opening %s!\n%s",progName,usage)
    return
  }
  defer prog.Close()
  progReader := bufio.NewReader(prog)

  for {
    line1, readErr1 := progReader.ReadString('\n')
    line1 = strings.Trim(line1,"\r\n")
    line1 = strings.Trim(line1,"\n")    
    if len(line1) > 0 {
      if strings.Contains(line1,"\t0\t0") {
      } else if strings.Contains(line1,"0\t0\t"){        
      } else if strings.Contains(line1,"seq0_leftend\tseq0_rightend"){
      } else if strings.Contains(line1,"-"){
        lineArray := strings.Split(line1,"\t")
        lineArray0, _ := strconv.Atoi(lineArray[0])
        lineArray1, _ := strconv.Atoi(lineArray[1])
        lineArray2, _ := strconv.Atoi(lineArray[2])
        lineArray3, _ := strconv.Atoi(lineArray[3])
        if lineArray0 < 0 && lineArray1 < 0 {
          lineArray0 = 0 - lineArray0
          lineArray1 = 0 - lineArray1
          lineArray2 = 0 - lineArray2
          lineArray3 = 0 - lineArray3
        }
        fmt.Printf("%d\t%d\t%d\t%d\n",lineArray0,lineArray1,lineArray2,lineArray3)
      } else {
        fmt.Printf("%s\n",line1)
      }
    }

    if readErr1 == io.EOF {
      break
    }
  } 
}
