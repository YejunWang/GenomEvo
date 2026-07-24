//orthoCombine.go: to merge the adjacent blocks within the same orthologous regions with a desired resolution
//USAGE: ./orthoCombine  orthoParsing_result RESOLUTION_SIZE > ORTHO_COMBINING_RESULTS

package orthoCombine

import (
  "fmt"
  "io"
  "bufio"
  "strings"
  "strconv"
  "os"
)

func Main() {
  usage := "\northoCombine: to merge the adjacent blocks within the same orthologous regions with a desired resolution.\n\n" + "USAGE: ./orthoCombine  orthoParsing_result RESOLUTION_SIZE > ORTHO_COMBINING_RESULTS\n\n"
  if len(os.Args) < 3 {
    fmt.Printf("%s",usage)
    return
  }
  orthBlkFileName := os.Args[1]
  res,_ := strconv.Atoi(os.Args[2])

  orthBlkFile, openErr := os.Open(orthBlkFileName)
  if openErr != nil {
    fmt.Printf("Error happened in opening %s!\n%s",orthBlkFileName,usage)
    return
  }
  defer orthBlkFile.Close()

  orthBlkFileReader := bufio.NewReader(orthBlkFile)
  var block1_st, block1_end, block2_st, block2_end, block1_end_0, blockLen int
  var ortholog, ortholog1, orthologA string
  block1_stMap := make(map[string]int)
  block2_stMap := make(map[string]int)
  block1_endMap := make(map[string]int)
  block2_endMap := make(map[string]int)
  flag := 0 
  for {
    line, readErr := orthBlkFileReader.ReadString('\n')
    line = strings.Trim(line,"\r\n")
    line = strings.Trim(line,"\n")    
    if len(line) > 0 {
      lineArray := strings.Split(line,"\t")
      block1_st,_ = strconv.Atoi(lineArray[0])
      block1_end,_ = strconv.Atoi(lineArray[1])
      block2_st,_ = strconv.Atoi(lineArray[3])
      block2_end,_ = strconv.Atoi(lineArray[4])
      ortholog = lineArray[8]
      if ortholog == "RC" || ortholog == "HOM" {
      	if flag == 1 {
      	  fmt.Printf("%d\t%d\t%d\t%d\t%s\n",block1_stMap[orthologA], block1_endMap[orthologA], block2_stMap[orthologA], block2_endMap[orthologA], orthologA)
      	  flag = 0	
      	}
        fmt.Printf("%s\t%s\t%s\t%s\t%s\n",lineArray[0],lineArray[1],lineArray[3],lineArray[4],ortholog)
      } else if ortholog != ortholog1 {
        if len(orthologA) > 0 && flag == 1 { 
          fmt.Printf("%d\t%d\t%d\t%d\t%s\n",block1_stMap[orthologA], block1_endMap[orthologA], block2_stMap[orthologA], block2_endMap[orthologA], orthologA)
        }
        orthologA = lineArray[0] + "_" + lineArray[1]
        block1_stMap[orthologA] = block1_st
        block1_endMap[orthologA] = block1_end
        block2_stMap[orthologA] = block2_st
        block2_endMap[orthologA] = block2_end
        flag = 1 
      } else if ortholog == ortholog1 {
        blockLen = block1_st - block1_end_0
        if blockLen >= res {
          fmt.Printf("%d\t%d\t%d\t%d\t%s\n",block1_stMap[orthologA], block1_endMap[orthologA], block2_stMap[orthologA], block2_endMap[orthologA], orthologA)
          orthologA = lineArray[0] + "_" + lineArray[1]
          block1_stMap[orthologA] = block1_st
          block1_endMap[orthologA] = block1_end
          block2_stMap[orthologA] = block2_st
          block2_endMap[orthologA] = block2_end
          flag = 1  
        } else if blockLen < res {
          block1_endMap[orthologA] = block1_end
          block2_endMap[orthologA] = block2_end
          flag = 1
        }
      }
      ortholog1 = ortholog
      block1_end_0 = block1_end       	
    }
    if readErr == io.EOF {
      if flag == 1 {
      	  fmt.Printf("%d\t%d\t%d\t%d\t%s\n",block1_stMap[orthologA], block1_endMap[orthologA], block2_stMap[orthologA], block2_endMap[orthologA], orthologA)
      }
      break
    }     
  }  
}  
