//combSingleGeneOrtho   <SINGLE_GENE_OTHOLOGS_1_VS_2>   <SINGLE_GENE_OTHOLOGS_1_VS_3>  >SINGLE_GENE_OTHOLOGS_1_VS_2_VS_3
package combsingle

import (
  "fmt"
  "os"
  "io"
  "bufio"
  "strings"
)

func RunCombSingle () {
	if len(os.Args) <= 1 {
	   fmt.Printf("No input file!\n")
	   return
	}
	inputFileName1 := os.Args[1]
  inputFileName2 := os.Args[2]

	inputFile1, openErr1 := os.Open(inputFileName1)
	if openErr1 != nil {
	  fmt.Printf("Error happened in opening %s!\n",inputFileName1)
	  return
	}
	defer inputFile1.Close()
  
  inputFile2, openErr2 := os.Open(inputFileName2)
  if openErr2 != nil {
    fmt.Printf("Error happened in opening %s!\n",inputFileName2)
    return
  }
  defer inputFile2.Close()

	inputReader1 := bufio.NewReader(inputFile1)
  inputReader2 := bufio.NewReader(inputFile2)

  orth1 := make(map[string]string)
	for {
	    line, readErr1 := inputReader1.ReadString('\n')
      line = strings.Trim(line,"\r\n")
      line = strings.Trim(line,"\n")     
      if len(line) > 0 {
        lineArray := strings.Split(line,"\t")
        orth1[lineArray[0]] = line
      }
      if readErr1 == io.EOF {
        break
      }
  }
  
  for {
      line, readErr2 := inputReader2.ReadString('\n')
      line = strings.Trim(line,"\r\n")
      line = strings.Trim(line,"\n")
      if len(line) > 0 {
        lineArray := strings.Split(line,"\t")
        if len(orth1[lineArray[0]]) > 0 {
          fmt.Printf("%s\t%s\n",orth1[lineArray[0]],lineArray[1])
        }
      }
      if readErr2 == io.EOF {
        break
      }    
  }
}