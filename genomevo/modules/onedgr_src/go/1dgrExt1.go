// usage: ./1dgrExt1  nA.vs.GenusAG.orthBlk.txt  nA_AG  Genus_AG
package main

import (
  "fmt"
  "io"
  "bufio"
  "strings"
  "os"
)

func main () {
  if len(os.Args) < 4 {
    fmt.Printf("Error happened in input file!\n")
    return
  }
  inputFileName := os.Args[1]

  inputFile, openErr := os.Open(inputFileName)
  if openErr != nil {
    fmt.Printf("Error happened in opening %s!\n",inputFileName)
    return
  }
  defer inputFile.Close()

  inputFileReader := bufio.NewReader(inputFile)
  for {
    line, readErr := inputFileReader.ReadString('\n')
    line = strings.Trim(line,"\r\n")
    line = strings.Trim(line,"\n")
    if len(line) > 0 {
      lineArray := strings.Split(line,"\t")
      fmt.Printf("%s\t%s-%s\t%s\n",os.Args[2],lineArray[0],lineArray[1],os.Args[3])
    }

    if readErr == io.EOF {
      break
    }   
  }
}
