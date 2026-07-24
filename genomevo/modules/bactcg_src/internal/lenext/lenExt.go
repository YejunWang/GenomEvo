/* lenExt:  
   Extracting the length of individual proteins in FASTA file
   
   USAGE: ./lenExt  PROTEOME_FASTA_FILE
*/

package lenext

import (
  "fmt"
  "os"
  "io"
  "bufio"
  "strings"
)

func RunLenExt () {
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
  var gene, gene0, seq string
  for {
    line, readErr := inputReader.ReadString('\n')
    line = strings.Trim(line,"\r\n")
    line = strings.Trim(line,"\n")
    if (strings.Contains(line,">")) {
      line = strings.Replace(line,">","",-1)
      lineArray1 := strings.Fields(line)
      gene0 = lineArray1[0]      
      if len(gene) > 0 {
        fmt.Printf("%s\t%d\n", gene,len(seq))
        seq = ""
      }
      gene = gene0  
    } else if len(line) != 0 {
      seq += line 
    }  
    if readErr == io.EOF {
      fmt.Printf("%s\t%d\n", gene,len(seq))
      break
    }
  }   
}
