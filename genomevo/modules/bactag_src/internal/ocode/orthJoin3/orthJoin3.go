//orthJoin3.go: Joining the ortholog blocks from two ORTH_JOINING_RESULT_FILES
//USAGE: ./orthJoin3  ORTH_JOINING_RESULT_FILE_1  ORTH_JOINING_RESULT_FILE_2

package orthJoin3

import (
  "fmt"
  "io"
  "bufio"
  "strings"
  "strconv"
  "os"
)

func Main() {
  usage := "orthJoin3: to join the ortholog blocks from two ORTH_JOINING_RESULT_FILES.\n\n" + "./orthJoin3  ORTH_JOINING_RESULT_FILE_1  ORTH_JOINING_RESULT_FILE_2 > ORTH_JOINING_RESULTS\n\n"
  if len(os.Args) < 3 {
    fmt.Printf("%s",usage)
    return
  }
  orthCombFile1Name := os.Args[1]
  orthCombFile2Name := os.Args[2] 

  orthCombFile1, openErr1 := os.Open(orthCombFile1Name)
  if openErr1 != nil {
    fmt.Printf("Error happened in opening %s!\n%s",orthCombFile1Name,usage)
    return
  }
  defer orthCombFile1.Close()
  orthCombFile1Reader := bufio.NewReader(orthCombFile1)

  hom := make(map[int]string)
  sp := make(map[int]string)
  var block1_st, block1_end, maxCoord int
  var ann, sp1, sp2 string
  for {
    line1, readErr1 := orthCombFile1Reader.ReadString('\n')
    line1 = strings.Trim(line1,"\r\n")
    line1 = strings.Trim(line1,"\n")    
    if len(line1) > 0 {
      lineArray1 := strings.Split(line1,"\t")
      block1_st,_ = strconv.Atoi(lineArray1[0])
      block1_end,_ = strconv.Atoi(lineArray1[1])
      ann = lineArray1[2]
      sp1 = lineArray1[3]

      for i:= block1_st; i<=block1_end; i++ {
        hom[i] = ann
        sp[i] = sp1
      }
    }
    if readErr1 == io.EOF {
      maxCoord = block1_end
      break
    }   
  }

  orthCombFile2, openErr2 := os.Open(orthCombFile2Name)
  if openErr2 != nil {
    fmt.Printf("Error happened in opening %s!\n%s",orthCombFile2Name,usage)
    return
  }
  defer orthCombFile2.Close()
  orthCombFile2Reader := bufio.NewReader(orthCombFile2)
  for {
    line2, readErr2 := orthCombFile2Reader.ReadString('\n')
    line2 = strings.Trim(line2,"\r\n")
    line2 = strings.Trim(line2,"\n")    
    if len(line2) > 0 {
      lineArray2 := strings.Split(line2,"\t")
      block1_st,_ = strconv.Atoi(lineArray2[0])
      block1_end,_ = strconv.Atoi(lineArray2[1])
      ann = lineArray2[2]
      sp2 = lineArray2[3]
      for i:= block1_st; i<=block1_end; i++ {
        if len(hom[i]) > 0 {
          if ann == "ORTH" && hom[i] !=  ann {
            hom[i] = ann
            sp[i] = sp2
          }
        } else {
          hom[i] = ann
          sp[i] = sp2
        }
      }
    }
    if readErr2 == io.EOF {
      if maxCoord <= block1_end {
        maxCoord = block1_end
      }
      break
    }
  }     
  
  var st, en int
  for i:=1;i<=maxCoord;i++ {
    if len(hom[i]) > 0 && !(len(hom[i-1]) > 0) {
      if st > 0 {
        fmt.Printf("%d\t%d\t%s\t%s\n",st,en,hom[st],sp[st])
      }
      st = i
      en = i
    } else if len(hom[i]) > 0 && len(hom[i-1]) > 0 {
      if hom[i] == hom[i-1] {
        en = i
      } else {
        fmt.Printf("%d\t%d\t%s\t%s\n",st,en,hom[st],sp[st])
        st = i
        en = i
      }
    } 
  }
  fmt.Printf("%d\t%d\t%s\t%s\n",st,en,hom[st],sp[st])
}   