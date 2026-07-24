//bestpicker
//USAGE: ./bestpicker BLAST_OUTFMT6 >BEST_HIT

package bestpicker

import (
  "fmt"
  "os"
  "io"
  "bufio"
  "strings"
  "strconv"
)

func RunBestPicker () {
  if len(os.Args) <= 1 {
    fmt.Printf("No imput file!\n")
    return
  }
  inputFileName := os.Args[1]
  inputFile, openErr := os.Open(inputFileName)
  if openErr != nil {
    fmt.Printf("Error happened in opening %s!\n",inputFileName)
    return
  }
  defer inputFile.Close()

  inputReader := bufio.NewReader(inputFile)
  ortho := make(map[string]string)
  for {
    line, readErr := inputReader.ReadString('\n')
    line = strings.Trim(line,"\r\n")
    line = strings.Trim(line,"\n")
    if len(line) != 0 {
      lineArray := strings.Split(line,"\t")
      sim,_ := strconv.ParseFloat(lineArray[2],32)
      rem := lineArray[3]
      for i:=4; i<len(lineArray);i++ {
        rem += "\t" + lineArray[i]
      }
      if len(ortho[lineArray[0]]) == 0 {
        ortho[lineArray[0]] = lineArray[1]
        fmt.Printf("%s\t%s\t%f\t%s\n",lineArray[0],lineArray[1],sim,rem)
      }
    }  
    if readErr == io.EOF {
      break
    }
  } 	
}