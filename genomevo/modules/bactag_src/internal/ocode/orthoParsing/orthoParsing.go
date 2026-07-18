//orthoParsing: to identify the orthologous backbone of two genomes from HomBlk file
//USAGE: ./orthoParsing  MAUVE_BACKGROUND_HOMBLK  > ORTHO_BLOCK_ANNOTATION_RESULTS

package orthoParsing

import (
  "fmt"
  "io"
  "bufio"
  "strings"
  "strconv"
  "os"
  "sort"
)

type Pair struct {
    Key string
    Value int
}

type PairList []Pair

func (p PairList) Swap(i, j int) { 
  p[i], p[j] = p[j], p[i]
}
func (p PairList) Len() int {
  return len(p) 
}
func (p PairList) Less(i, j int) bool {
  return p[i].Value < p[j].Value 
}

func sortMapByValue(m map[string]int) PairList {
  p := make(PairList, len(m))
  i := 0
  for k, v := range m {
    p[i] = Pair{k, v}
    i++
  }
  sort.Sort(p)
  return p
}

func Main() {
  usage := "\northoParsing: to identify the orthologous backbone of two genomes from HomBlk file. \n\n" + "USAGE: ./orthoParsing  MAUVE_BACKGROUND_HOMBLK\n\n"
  if len(os.Args) < 2 {
    fmt.Printf("%s",usage)
    return
  }
  homBlkFileName := os.Args[1]

  homBlkFile, openErr := os.Open(homBlkFileName)
  if openErr != nil {
    fmt.Printf("Error happened in opening %s!\n%s",homBlkFileName,usage)
    return
  }
  defer homBlkFile.Close()

  homBlkFileReader := bufio.NewReader(homBlkFile)
  
  flagNewBlock := 1; flagContBlock := 0; lineCount := 0
  hom := make(map[string]string)
  block := make(map[string]string)
  homBlkLen := make(map[string]int)
  lineStart := make(map[string]int)
  blockLenSort := make(map[int]string)
  var blockName string 
  var homBlk_1, homBlk_0, lineAll int
  for {
    line1, readErr1 := homBlkFileReader.ReadString('\n')
    line1 = strings.Trim(line1,"\r\n")
    line1 = strings.Trim(line1,"\n")
    if len(line1) > 0 {
      lineCount++
      lineArray1 := strings.Split(line1,"\t")
      homBlk2RankStr := strings.Replace(lineArray1[5],"2_HomBlock_","",-1)
      homBlk2Rank,_ := strconv.Atoi(homBlk2RankStr)
      homBlk1St,_ := strconv.Atoi(lineArray1[0])
      homBlk1End,_ := strconv.Atoi(lineArray1[1]) 
//      homBlk1End,_ := strconv.ParseInt(lineArray1[1],10,0)      
      if strings.Contains(lineArray1[3],"-") {
        hom[lineArray1[0]] = "RC"
        flagNewBlock = 0
      } else if flagNewBlock == 0 {
        flagNewBlock = 1
        flagContBlock = 1
        block[lineArray1[0]] = lineArray1[5]
        blockName = lineArray1[5]
        homBlk_1 = homBlk2Rank
        homBlkLen[lineArray1[5]] = homBlk1End - homBlk1St + 1
        lineStart[lineArray1[5]] = lineCount
      } else if (flagNewBlock == 1 && flagContBlock == 0) {
      	flagContBlock = 1
        block[lineArray1[0]] = lineArray1[5]
        blockName = lineArray1[5]
        homBlk_1 = homBlk2Rank
        homBlkLen[lineArray1[5]] = homBlk1End - homBlk1St + 1
        lineStart[lineArray1[5]] = lineCount   	
      } else if (flagNewBlock == 1 && flagContBlock == 1) {
        homBlk_0 = homBlk2Rank
        if homBlk_0 - homBlk_1 == 1 {
          block[lineArray1[0]] = blockName
          homBlk_1 = homBlk_0         
          homBlkLen[blockName] += homBlk1End - homBlk1St + 1	
        } else {
          block[lineArray1[0]] = lineArray1[5]
          blockName = lineArray1[5]
          homBlk_1 = homBlk2Rank         
          homBlkLen[blockName] = homBlk1End - homBlk1St + 1
          lineStart[lineArray1[5]] = lineCount
        }
      }
    }
    if readErr1 == io.EOF {
      lineAll = lineCount
      var sortBlkLenPair PairList
      sortBlkLenPair = sortMapByValue(homBlkLen)
      lenRank := 0
      for _, o := range sortBlkLenPair {
      	lenRank++
      	lenRankRev := len(homBlkLen) - lenRank + 1
        blockLenSort[lenRankRev] = o.Key
      }
    //  for k,v := range blockLenSort {
    //  	fmt.Printf("%d\t%s\n",k,v)
    //  }
      break
    }    
  }

  lineCount = 0
  new1Block := make(map[string]int)
  newBlock1Rank := make(map[string]int)
  new2Block := make(map[string]int)
  newBlock2Rank := make(map[string]int)
  longestBlk2St, _ :=  strconv.Atoi(strings.Replace(blockLenSort[1],"2_HomBlock_","",-1))
  homBlkFile, openErr = os.Open(homBlkFileName)
  if openErr != nil {
    fmt.Printf("Error happened in opening %s!\n%s",homBlkFileName,usage)
    return
  }
  defer homBlkFile.Close()
  homBlkFileReader = bufio.NewReader(homBlkFile)
  for {
    line2, readErr2 := homBlkFileReader.ReadString('\n')
    line2 = strings.Trim(line2,"\r\n")
    line2 = strings.Trim(line2,"\n")
    if len(line2) > 0 {
      lineCount++
      lineArray2 := strings.Split(line2,"\t")
      homBlk2RankStr := strings.Replace(lineArray2[5],"2_HomBlock_","",-1)
      homBlk2Rank,_ := strconv.Atoi(homBlk2RankStr)      
      if lineCount < lineStart[blockLenSort[1]] {
        new1Block[lineArray2[0]] = lineAll - (lineStart[blockLenSort[1]] - 1) + lineCount
        if homBlk2Rank - longestBlk2St + 1 < 0 {
          new2Block[lineArray2[0]] = lineAll - (longestBlk2St - 1) + homBlk2Rank
        } else {
          new2Block[lineArray2[0]] = homBlk2Rank - (longestBlk2St - 1)
        }
      } else {
        new1Block[lineArray2[0]] = lineCount -  (lineStart[blockLenSort[1]] - 1)
        if homBlk2Rank - longestBlk2St + 1 < 0 {
          new2Block[lineArray2[0]] = lineAll - (longestBlk2St - 1) + homBlk2Rank
        } else {
          new2Block[lineArray2[0]] = homBlk2Rank - (longestBlk2St - 1)
        }
      }
      if len(block[lineArray2[0]]) > 0 {
        newBlock1Rank[block[lineArray2[0]]] = new1Block[lineArray2[0]]
        newBlock2Rank[block[lineArray2[0]]] = new2Block[lineArray2[0]]
      }  
    }
    if readErr2 == io.EOF {
    //  for k,v := range new1Block {
    //      fmt.Printf("%s\t%d\t%d\n",k,v,new2Block[k]) 
    //  }	
      break
    }  
  }
  
  ortho := make(map[string]int)
  var up1Diff, down1Diff, iMax, downBlock1Rank, upBlock1Rank, blockOrder1Diff int 
  for i:=1;i<=2;i++ {
  	ortho[blockLenSort[i]] = 1
  }
  for i:=3;i<=len(blockLenSort);i++ {
    up1Diff = -100000; down1Diff = 100000; iMax = 0
    for k:=1;k<i;k++ {
      if ortho[blockLenSort[k]] == 1 {
        blockOrder1Diff =  newBlock1Rank[blockLenSort[k]] -  newBlock1Rank[blockLenSort[i]]; 
        if (blockOrder1Diff > 0) && (blockOrder1Diff < down1Diff) {
          down1Diff = blockOrder1Diff
          downBlock1Rank = k
        } else if (blockOrder1Diff < 0) && (blockOrder1Diff > up1Diff) {
          up1Diff = blockOrder1Diff
          upBlock1Rank = k
        }
      }
    }
    if down1Diff == 100000 {
      iMax = 1
    }
    if iMax == 0 {
      if (newBlock2Rank[blockLenSort[i]] - newBlock2Rank[blockLenSort[upBlock1Rank]]) > 0 && (newBlock2Rank[blockLenSort[i]] - newBlock2Rank[blockLenSort[downBlock1Rank]]) < 0 {
        ortho[blockLenSort[i]] = 1
      } else {
        ortho[blockLenSort[i]] = 0
      }
    } else {
      if (newBlock2Rank[blockLenSort[i]] - newBlock2Rank[blockLenSort[upBlock1Rank]]) > 0 {
        ortho[blockLenSort[i]] = 1
      } else {
        ortho[blockLenSort[i]] = 0
      }
    }
    // fmt.Printf("%d\t%s\t%d\n",i,blockLenSort[i],ortho[blockLenSort[i]])
  }

  homBlkFile, openErr = os.Open(homBlkFileName)
  if openErr != nil {
    fmt.Printf("Error happened in opening %s!\n%s",homBlkFileName,usage)
    return
  }
  defer homBlkFile.Close()
  homBlkFileReader = bufio.NewReader(homBlkFile)
  for {
    line3, readErr3 := homBlkFileReader.ReadString('\n')
    line3 = strings.Trim(line3,"\r\n")
    line3 = strings.Trim(line3,"\n")
    if len(line3) > 0 {
      lineArray3 := strings.Split(line3,"\t")
      if len(hom[lineArray3[0]]) > 0 {
      	fmt.Printf("%s\t%d\t%d\t%s\n",line3,new1Block[lineArray3[0]],new2Block[lineArray3[0]],hom[lineArray3[0]])
      } else if ortho[block[lineArray3[0]]] == 0 {
      	fmt.Printf("%s\t%d\t%d\tHOM\n",line3,new1Block[lineArray3[0]],new2Block[lineArray3[0]])      	
      } else {
      	fmt.Printf("%s\t%d\t%d\t%s\n",line3,new1Block[lineArray3[0]],new2Block[lineArray3[0]],block[lineArray3[0]])
      }
    }
    if readErr3 == io.EOF {
       break
    }
  }    
}



