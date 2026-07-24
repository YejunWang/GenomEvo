//mutbest  BEST_HIT_FILTER_1  BEST_HIT_FILTER_2 >BEST_HIT_FILTER_MUT

package mutbest

import (
  "fmt"
  "os"
  "io"
  "bufio"
  "strings"
)

func RunMutBest () {
  if len(os.Args) <= 2 {
    fmt.Printf("Two imput files required!\n")
    return
  }
  inputFileName1 := os.Args[1]
  inputFile1, openErr1 := os.Open(inputFileName1)
  if openErr1 != nil {
    fmt.Printf("Error happened in opening %s!\n",inputFileName1)
    return
  }
  defer inputFile1.Close()

  inputFileName2 := os.Args[2]
  inputFile2, openErr2 := os.Open(inputFileName2)
  if openErr2 != nil {
    fmt.Printf("Error happened in opening %s!\n",inputFileName2)
    return
  }
  defer inputFile2.Close()

  inputReader1 := bufio.NewReader(inputFile1)
  inputReader2 := bufio.NewReader(inputFile2)
  ortho := make(map[string]string)
  for {
    line, readErr1 := inputReader1.ReadString('\n')
    line = strings.Trim(line,"\r\n")
    line = strings.Trim(line,"\n")
    if len(line) != 0 && !(strings.Contains(line,"Similarity")){
        lineArray := strings.Split(line,"\t")
        ortho[lineArray[0]] = lineArray[1]
    //  sim[lineArray[0]] = lineArray[2]
    }  
    if readErr1 == io.EOF {
      break
    }
  } 

  
  for {
    line, readErr2 := inputReader2.ReadString('\n')
    line = strings.Trim(line,"\r\n")
    line = strings.Trim(line,"\n")
    if len(line) != 0 && !(strings.Contains(line,"Similarity")) {      
        lineArray := strings.Split(line,"\t")
        if len(ortho[lineArray[1]]) > 0 && ortho[lineArray[1]] == lineArray[0] {
          fmt.Printf("%s\t%s\n",lineArray[1],lineArray[0]) 
        }
    }  
    if readErr2 == io.EOF {
      break
    }
  } 
}