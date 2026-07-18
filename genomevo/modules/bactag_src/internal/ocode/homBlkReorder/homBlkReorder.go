//homBlkReorder.go	MAUVE_BACKBONE	>HOM_BLOCK_REORDER
package homBlkReorder

import (
  "fmt"
  "os"
  "io"
  "bufio"
  "strings"
  "strconv"
  "sort"
)

func Main() {
  usage := "homBlkReorder: extracting the coordinates of orthologs from Mauve backbone files.\n\n" + "USAGE:\nnonhom  MAUVE_BACKBONE  >ORTHO_COORDINATES\n\n"
  if len(os.Args) < 2 {
    fmt.Printf("%s",usage)
    return
  } 
  inputfileName := os.Args[1]
  inputfile, openErr := os.Open(inputfileName)
  if openErr != nil {
    fmt.Printf("Error happened in opening the file %s!\n",inputfileName)
    fmt.Printf("%s",usage)
    return
  }
  defer inputfile.Close()

  inputReader := bufio.NewReader(inputfile)
  slice1 := []int{}
  slice2 := []int{}
  slice2a := []int{}
  slice2b := []int{}
  en1Map := make(map[int]int)
  en2Map := make(map[int]int)
  homMap := make(map[int]int)
  ord := make(map[int]int)
  str := make(map[int]string)

  var st1, st2, en1, en2 int
  for {
    line, readErr := inputReader.ReadString('\n')
    line = strings.Trim(line, "\r\n")
    line = strings.Trim(line, "\n")    
    if len(line) > 0 {
        lineArray := strings.Split(line,"\t")    
        st1,_ = strconv.Atoi(lineArray[0])
        en1,_ = strconv.Atoi(lineArray[1])
        if strings.Contains(lineArray[2],"-") {
          st2,_ = strconv.Atoi(strings.Replace(lineArray[2],"-","",-1))
          en2,_ = strconv.Atoi(strings.Replace(lineArray[3],"-","",-1))
          str[st2] = "-"
        } else {
          st2,_ = strconv.Atoi(lineArray[2])
          en2,_ = strconv.Atoi(lineArray[3])
          str[st2] = "+"      
        }
        if st1 > en1 {
          st1, en1 = en1, st1
        }
        if st2 > en2 {
          st2, en2 = en2, st2
        }
        en1Map[st1] = en1
        en2Map[st2] = en2
        homMap[st1] = st2
        slice1 = append(slice1, st1) 
        slice2 = append(slice2, st2) 
    } 
    if readErr == io.EOF {
      sort.Ints(slice1)
      sort.Ints(slice2)
      for i := 0; i < len(slice2); i++ {
        if slice2[i] >= homMap[slice1[0]] {
          slice2a = append(slice2a, slice2[i])
        } else {
          slice2b = append(slice2b, slice2[i])
        }
      }    
      slice2 = slice2a 
      for i := 0; i < len(slice2b); i++ {
        slice2 = append(slice2, slice2b[i])
      }
      for i := 0; i < len(slice2); i++ {
        ord[slice2[i]] = i+1
      }
      for i := 0; i < len(slice1); i++ {
        if strings.Contains(str[homMap[slice1[i]]],"-") {
          fmt.Printf("%d\t%d\t1_HomBlock_%d\t-%d\t-%d\t2_HomBlock_%d\n",slice1[i],en1Map[slice1[i]],i+1,homMap[slice1[i]],en2Map[homMap[slice1[i]]],ord[homMap[slice1[i]]])
        } else {
          fmt.Printf("%d\t%d\t1_HomBlock_%d\t%d\t%d\t2_HomBlock_%d\n",slice1[i],en1Map[slice1[i]],i+1,homMap[slice1[i]],en2Map[homMap[slice1[i]]],ord[homMap[slice1[i]]])          
        }   
      }  
      break
    }
  }
}