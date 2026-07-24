//patching.go: to complement orthologous groups to the ortholog backbone
//USAGE: ./patching  TO_BE_COMPLEMENTED_ORTHOBLK_FILE  TO_COMPLEMENT_ORTHOBLK_FILE  BACKBONE_JOINT_ORTH_FILE  SOURCE_SP  HOM_SP ANNOTATION >HOM_SP_COMPLEMENTED_ORTH_FILE

/*

  Prepare the BACKBONE_JOINT_ORTH_FILE like:
  Source_Strain   Source_st   Source_end    Annotation    Hom_strain
  RKS3013         5           163191        ORTH          RKS3027
  RKS3013         166854      186068        ORTH          BAA-1581
  RKS3013         187174      346027        ORTH          BAA-1581
  RKS3013         352900      357780        ORTH          RKS3027
  
  NOTE that all "HOM" or "RC" blocks were removed, and the Source_st was ordered. 
  TO_BE_COMPLEMENTED_ORTHOBLK_FILE and TO_COMPLEMENT_ORTHOBLK_FILE could have "HOM" or "RC" blocks (not merged files).

*/
package patching

import (
  "fmt"
  "sort"
  "io"
  "bufio"
  "strings"
  "strconv"
  "os"
)

func Main() {
  usage := "patching: to complement orthologous groups to the ortholog backbone.\n\n" + "./patching  TO_BE_COMPLEMENTED_ORTHOBLK_FILE  TO_COMPLEMENT_ORTHOBLK_FILE  BACKBONE_JOINT_ORTH_FILE  SOURCE_SP  HOM_SP ANNOTATION >HOM_SP_COMPLEMENTED_ORTH_FILE\n\n"
  if len(os.Args) < 7 {
    fmt.Printf("%s",usage)
    return
  }
  orthBlkFile1Name := os.Args[1]
  orthBlkFile2Name := os.Args[2]
  backboneFileName := os.Args[3]
  sourceSp := os.Args[4]
  compSp := os.Args[5] 
  annot := os.Args[6]

// *** Determine the coordinates of blocks within source strain not covered by the backbone ancient genome *** // 

  orthBlkFile1, openErr1 := os.Open(orthBlkFile1Name)
  if openErr1 != nil {
    fmt.Printf("Error happened in opening %s!\n%s",orthBlkFile1Name,usage)
    return
  }
  defer orthBlkFile1.Close()
  orthBlkFile1Reader := bufio.NewReader(orthBlkFile1)

  lineCount := 0
  var span1_st, span2_st, span1, span2, block1_st, block1_end, block2_st, block2_end int
  var ann string
  toBeComp1_st := make(map[int]int)
  toBeComp1_end := make(map[int]int)
  toBeComp2_st := make(map[int]int)
  toBeComp2_end := make(map[int]int)
  block1DupRemove := make(map[int]int)
  block2DupRemove := make(map[int]int)
  for {
    line1, readErr1 := orthBlkFile1Reader.ReadString('\n')
    line1 = strings.Trim(line1,"\r\n")
    line1 = strings.Trim(line1,"\n")
    if len(line1) > 0 {
      lineArray1 := strings.Split(line1,"\t")
      block1_st,_ = strconv.Atoi(lineArray1[0])
      block1_end,_ = strconv.Atoi(lineArray1[1])
      if block1DupRemove[block1_st] == 1 {
        block1_st += 1
      }
      block2_st,_ = strconv.Atoi(lineArray1[3])
      block2_end,_ = strconv.Atoi(lineArray1[4])
      if block2DupRemove[block2_st] == 1 {
        block2_st += 1
      }
      block1DupRemove[block1_end] = 1
      block2DupRemove[block2_end] = 1     
      ann = lineArray1[8]
      if ann != "RC" && ann != "HOM" {
        if span1_st > 0 {
          span1 = block1_st - 1 - span1_st + 1
          span2 = block2_st - 1 - span2_st + 1
          if span2 == 0 && span1 > 1000 {
            lineCount++
            toBeComp1_st[lineCount] = span1_st - 1
            toBeComp1_end[lineCount] = block1_st
            toBeComp2_st[lineCount] = span2_st - 1
            toBeComp2_end[lineCount] = block2_st
    //        fmt.Printf("%d\t%d\t%d\t%d\n",toBeComp1_st[lineCount],toBeComp1_end[lineCount],toBeComp2_st[lineCount],toBeComp2_end[lineCount])            
          } else if span2 != 0 && span1 != 0 && span2 > -10000 { 
            spanRatio := float32(span2)/float32(span1)
            if span1 > 1000 && spanRatio < 0.1 {
              lineCount++
              toBeComp1_st[lineCount] = span1_st
              toBeComp1_end[lineCount] = block1_st - 1
              toBeComp2_st[lineCount] = span2_st
              toBeComp2_end[lineCount] = block2_st - 1
              if span2 < 0 {
                toBeComp2_st[lineCount] = toBeComp2_st[lineCount] - 1
                toBeComp2_end[lineCount] = toBeComp2_end[lineCount] + 2
              }
  //            fmt.Printf("%d\t%d\t%d\t%d\n",toBeComp1_st[lineCount],toBeComp1_end[lineCount],toBeComp2_st[lineCount],toBeComp2_end[lineCount])             
            }
          }  
        }
      	span1_st = block1_end + 1
      	span2_st = block2_end + 1
      }
    }
    if readErr1 == io.EOF {
      if ann == "RC" || ann == "HOM" && span1_st > 0 {
        span1 = block1_end - span1_st + 1
        span2 = block2_end - span2_st + 1
        if span2 == 0 && span1 > 1000 {
          lineCount++
          toBeComp1_st[lineCount] = span1_st - 1
          toBeComp1_end[lineCount] = block1_st
          toBeComp2_st[lineCount] = span2_st - 1
          toBeComp2_end[lineCount] = block2_st           
        } else if span2 != 0 && span1 != 0 && span2 > -10000 { 
          spanRatio := float32(span2)/float32(span1)
          if span1 > 1000 && spanRatio < 0.1 {
            lineCount++
            toBeComp1_st[lineCount] = span1_st
            toBeComp1_end[lineCount] = block1_st - 1
            toBeComp2_st[lineCount] = span2_st
            toBeComp2_end[lineCount] = block2_st - 1
            if span2 < 0 {
              toBeComp2_st[lineCount] = toBeComp2_st[lineCount] - 1
              toBeComp2_end[lineCount] = toBeComp2_end[lineCount] + 2
            }           
          }
        }
  //      fmt.Printf("%d\t%d\t%d\t%d\n",toBeComp1_st[lineCount],toBeComp1_end[lineCount],toBeComp2_st[lineCount],toBeComp2_end[lineCount])   
      }      
      break
    }
  }
//  fmt.Printf("\n######################################\n")

// *** Determine the coordinates of insert sites within the backbone ancient genome with complemented ancient genome block sequences within source strain *** //

/* ###########################################################
 order  comp1_st      comp1_end       comp2_st      comp2_end
  1     74329         82826           74329         74330
  2     163058        170695          154561        154562
  3     203834        225127          187700        187701
############################################################## */ 

  orthBlkFile2, openErr2 := os.Open(orthBlkFile2Name)
  if openErr2 != nil {
    fmt.Printf("Error happened in opening %s!\n%s",orthBlkFile2Name,usage)
    return
  }
  defer orthBlkFile2.Close()
  orthBlkFile2Reader := bufio.NewReader(orthBlkFile2)
  
  count := 0
  comp1_st := make(map[int]int)
  comp1_end := make(map[int]int)
  comp2_st := make(map[int]int)
  comp2_end := make(map[int]int)  
  for {
    line2, readErr2 := orthBlkFile2Reader.ReadString('\n')
    line2 = strings.Trim(line2,"\r\n")
    line2 = strings.Trim(line2,"\n")
    if len(line2) > 0 {
      lineArray2 := strings.Split(line2,"\t")
      block1_st,_ = strconv.Atoi(lineArray2[0])
      block1_end,_ = strconv.Atoi(lineArray2[1])
      block2_st,_ = strconv.Atoi(lineArray2[3])
      block2_end,_ = strconv.Atoi(lineArray2[4])      
      ann = lineArray2[8]
      if ann != "RC" && ann != "HOM" {
        for i:=1;i<=len(toBeComp1_st);i++ {
          //block1_st   toBeComp1_st  toBeComp1_end  block1_end
          //block1_st   toBeComp1_st  block1_end toBeComp1_end  
          //toBeComp1_st  block1_st  toBeComp1_end  block1_end
          //toBeComp1_st  block1_st  block1_end  toBeComp1_end          
          if block1_st <= toBeComp1_st[i] && block1_end >= toBeComp1_end[i] {
            count++
            comp1_st[count] = toBeComp1_st[i]
            comp1_end[count] = toBeComp1_end[i]
            comp2_st[count] = toBeComp2_st[i]
            comp2_end[count] = toBeComp2_end[i]
  //          fmt.Printf("%d\t%d\t%d\t%d\t%d\n",count,comp1_st[count],comp1_end[count],comp2_st[count],comp2_end[count])
          } else if block1_st <= toBeComp1_st[i] && toBeComp1_st[i] <= block1_end && block1_end < toBeComp1_end[i] {
            if block1_end - toBeComp1_st[i] + 1 >= 1000 {
              count++
              comp1_st[count] = toBeComp1_st[i]
              comp1_end[count] = block1_end
              comp2_st[count] = toBeComp2_st[i]
              comp2_end[count] = toBeComp2_end[i]
  //            fmt.Printf("%d\t%d\t%d\t%d\t%d\n",count,comp1_st[count],comp1_end[count],comp2_st[count],comp2_end[count])              
            }
          } else if block1_st > toBeComp1_st[i] && block1_st <= toBeComp1_end[i] && toBeComp1_end[i] <= block1_end {
            if toBeComp1_end[i] - block1_st + 1 >= 1000 {
              count++
              comp1_st[count] = block1_st
              comp1_end[count] = toBeComp1_end[i]
              comp2_st[count] = toBeComp2_st[i]
              comp2_end[count] = toBeComp2_end[i]
  //           fmt.Printf("%d\t%d\t%d\t%d\t%d\n",count,comp1_st[count],comp1_end[count],comp2_st[count],comp2_end[count])              
            }
          } else if block1_st > toBeComp1_st[i] &&  toBeComp1_end[i] > block1_end {
            if block1_end - block1_st + 1 >= 1000 {
              count++
              comp1_st[count] = block1_st
              comp1_end[count] = block1_end
              comp2_st[count] = toBeComp2_st[i]
              comp2_end[count] = toBeComp2_end[i]
  //            fmt.Printf("%d\t%d\t%d\t%d\t%d\n",count,comp1_st[count],comp1_end[count],comp2_st[count],comp2_end[count])              
            }
          }
        }
      }  
    }
    if readErr2 == io.EOF {
      break
    }    
  }

//fmt.Printf("\n######################################\n")

// *** Combining the multiple complemented blocks for the same inserted sites *** //

/* ####################################################################################
   Case1: realBG    ---------------
          comp2         -----
   Case2: realBG    ---------------
          comp2     -----
   Case3: realBG    ---------------
          comp2               -----
   Case4: realBG    ---------------      -----------------     --------------
          comp2   --                    --                    ---------------
   Case5: realBG    ---------------
          comp2               --------
   Case6: realBG    ---------------      -----------------       --------------
          comp2   -----------------      ------------------     -----------------         
####################################################################################### */ 

  var j string
  mulFlag := make(map[int]int)
  newMap2 := make(map[int]int)
  newMap3 := make(map[int]int)
  comp := make(map[string]string)
  comp1 := make(map[int]string)
  mulIns := make(map[string]int)
  newMap1 := make(map[int]string)

  for i:=1;i<=len(comp2_st);i++{
    str := strconv.Itoa(comp2_st[i]) + "_" + strconv.Itoa(comp2_end[i])
    if mulIns[str] == 1 {
      mulFlag[i] = 1
      j += "\t" + strconv.Itoa(i)
      comp[str] += sourceSp + "\t" + strconv.Itoa(comp1_st[i]) + "\t" + strconv.Itoa(comp1_end[i]) + "\t" + annot + "\t" + compSp + "\n"
      comp1[i] = comp[str]
    } else {
      if len(j) > 0 {
        if strings.Contains(j,"\t") {
          jArray := strings.Split(j,"\t")
          l := i - 1
          for k:=0;k<len(jArray);k++ {
            if len(jArray[k]) > 0 {
              m,_ := strconv.Atoi(jArray[k])
              comp1[m] = comp1[l]
              //fmt.Printf("%s",comp1[m])
            }
          }
        }
      }
      j = strconv.Itoa(i)
      comp[str] =  sourceSp + "\t" + strconv.Itoa(comp1_st[i]) + "\t" + strconv.Itoa(comp1_end[i]) + "\t" + annot + "\t" + compSp + "\n"
      comp1[i] = comp[str] 
      mulIns[str] = 1
    }
  }
  kmax := 0
  if strings.Contains(j,"\t") {
    jArray := strings.Split(j,"\t")
    for k:=0;k<len(jArray);k++ {
      n,_ := strconv.Atoi(jArray[k])
      if len(jArray[k]) > 0 && kmax < n {
        kmax = n
      }
    }
    for k:=0;k<len(jArray);k++ {
      if len(jArray[k]) > 0 {
        m,_ := strconv.Atoi(jArray[k])
        comp1[m] = comp1[kmax]
      }
    }
  }

/*
for ke, va := range comp1 {
  fmt.Println(ke,va,strconv.Itoa(comp2_st[ke]),"\n",strconv.Itoa(comp2_end[ke]))
}
*/
// 1. 提取 comp1 的键到切片 keys 中
keys := make([]int, 0, len(comp1))
for k := range comp1 {
    keys = append(keys, k)
}

// 2. 按照 comp2_st[ke] 和 comp2_end[ke] 的值对 keys 进行排序
sort.Slice(keys, func(i, j int) bool {
    if comp2_st[keys[i]] != comp2_st[keys[j]] {
        return comp2_st[keys[i]] < comp2_st[keys[j]]
    }
    // 如果 comp2_st 相等，则按 comp2_end 排序
    return comp2_end[keys[i]] < comp2_end[keys[j]]
})

// 3. 建立一个 map 用来记录每个值第一次出现的键
valueToFirstKey := make(map[string]int)
for _, ke := range keys {
    if _, ok := valueToFirstKey[comp1[ke]]; !ok {
        valueToFirstKey[comp1[ke]] = ke
    }
}

// 4. 输出符合条件的键值对到 newMap1, newMap2, newMap3


for idx, ke := range keys {
    va := comp1[ke]
    // 只记录每个值第一次出现的键值对
    if ke == valueToFirstKey[va] {
        newMap1[idx+1] = va
        newMap2[idx+1] = comp2_st[ke]
        newMap3[idx+1] = comp2_end[ke]
    }
}

// 5. 输出 newMap1 的内容，剔除值相同的键值对
/*
for k, v := range newMap1 {
    fmt.Println(k, v, strconv.Itoa(newMap2[k]), "\n", strconv.Itoa(newMap3[k]))
}
*/




  backboneFile, openErr3 := os.Open(backboneFileName)
  if openErr3 != nil {
    fmt.Printf("Error happened in opening %s!\n%s",backboneFileName,usage)
    return
  }
  defer backboneFile.Close()
  backboneFileReader := bufio.NewReader(backboneFile)
  realBGst := 1
  realBGend := 0
  flag := 0
  flag1 := 0
//  var x string
//  var y int


// *** Complementing the backbone ortholog block file *** //

/* ####################################################################################
 order  source_str  comp1_st    comp1_end  comp2_st       hom_str   comp2_st  comp2_end
   1      LT2       74329       82826      nD_AG_added    nD_AG     74329     74330
   2      LT2       203834      225127     nD_AG_added    nD_AG     187700    187701
   3      LT2       1090902     1092428    nD_AG_added    nD_AG     1041123   1041124
   4      LT2       3566638     3568341    nD_AG_added    nD_AG
          LT2       3570128     3572990    nD_AG_added    nD_AG     3422268   3422269
   5      LT2       3566638     3568341    nD_AG_added    nD_AG
          LT2       3570128     3572990    nD_AG_added    nD_AG     3422268   3422269
####################################################################################### */ 


var output string
flag3 := make(map[int]int)
for {
  line3, readErr3 := backboneFileReader.ReadString('\n')
  line3 = strings.Trim(line3,"\r\n")
  line3 = strings.Trim(line3,"\n")
  if len(line3) > 0 {
	lineArray3 := strings.Split(line3,"\t")
	bgSourceSp := lineArray3[0]
	source_st,_ := strconv.Atoi(lineArray3[1])
	source_end,_ := strconv.Atoi(lineArray3[2])
	realBGend = source_end - source_st + realBGst
	bgHomSp := lineArray3[4]
	ann := lineArray3[3]
	for i:=1;i<=len(newMap2);i++ {
	  if mulFlag[i] != 1 && flag3[i] != 1 {
		a := newMap2[i] - realBGst + source_st
		b := newMap3[i] - realBGst + source_st
		if realBGst < newMap2[i] && newMap3[i] < realBGend {
		  flag3[i] = 1
		  if flag == 0 {
			 output = bgSourceSp + "\t" + strconv.Itoa(source_st) + "\t" + strconv.Itoa(a) + "\t" + ann + "\t" + bgHomSp + "\n"
			 output += newMap1[i]
			 output += bgSourceSp + "\t" + strconv.Itoa(b) + "\t"  //source_end,ann,bgHomSp
  //           x = "Case1-1"
  //           y = i
  //           fmt.Printf("%s\t%s\t%d\t%s\n",line3,x,y,newMap1[y])
		  } else if flag == 1 {
			output += strconv.Itoa(a) + "\t" + ann + "\t" + bgHomSp + "\n"
			output += newMap1[i]
			output += bgSourceSp + "\t" + strconv.Itoa(b) + "\t"
  //          x = "Case1-2"
  //          y = i
  //          fmt.Printf("%s\t%s\t%d\t%s\n",line3,x,y,newMap1[y]) 
		  }
		  flag = 1 
		} else if realBGst == newMap2[i] && newMap3[i] < realBGend {
		  flag3[i] = 1
		  output = newMap1[i]
  //        x = "Case2"
  //        y = i
  //        fmt.Printf("%s\t%s\t%d\t%s\n",line3,x,y,newMap1[y])
		  if b - a == 1 {
			output += bgSourceSp + "\t" + strconv.Itoa(a) + "\t" //source_end,ann,bgHomSp
		  } else {
			output += bgSourceSp + "\t" + strconv.Itoa(b) + "\t" //source_end,ann,bgHomSp
		  }
		  flag = 1
		} else if realBGst < newMap2[i] && newMap3[i] == realBGend{
		  flag3[i] = 1
		  if flag == 0 {
			output = bgSourceSp + "\t" + strconv.Itoa(source_st) + "\t" + strconv.Itoa(a) + "\t" + ann + "\t" + bgHomSp + "\n"
			output += newMap1[i]
			flag1 = 1
  //          x = "Case3-1"
  //          y = i
  //          fmt.Printf("%s\t%s\t%d\t%s\n",line3,x,y,newMap1[y])
		  } else if flag == 1 {
			output += strconv.Itoa(a) + "\t" + ann + "\t" + bgHomSp + "\n"
			output += newMap1[i]
			flag1 = 1
  //          x = "Case3-2"
  //          y = i
  //          fmt.Printf("%s\t%s\t%d\t%s\n",line3,x,y,newMap1[y])
		  }
		  flag = 1
		} else if newMap2[i] < realBGst && realBGst <= newMap3[i] && newMap3[i] <= realBGend {
		  flag3[i] = 1
		  output = newMap1[i]
		  output += bgSourceSp + "\t" + strconv.Itoa(b) + "\t" //source_end,ann,bgHomSp
		  flag = 1 
  //        x = "Case4"
  //        y = i
  //        fmt.Printf("%s\t%s\t%d\t%s\n",line3,x,y,newMap1[y])          
		} else if realBGst < newMap2[i] && newMap2[i] < realBGend && newMap3[i] > realBGend {
		  flag3[i] = 1
		  if flag == 0 {
			output = bgSourceSp + "\t" + strconv.Itoa(source_st) + "\t" + strconv.Itoa(a) + "\t" + ann + "\t" + bgHomSp + "\n"
			output += newMap1[i]
			flag1 = 1
  //          x = "Case5-1"
  //          y = i
  //          fmt.Printf("%s\t%s\t%d\t%s\n",line3,x,y,newMap1[y])
		  } else if flag == 1 {
			output += strconv.Itoa(a) + "\t" + ann + "\t" + bgHomSp + "\n"
			output += newMap1[i]
			flag1 = 1
  //          x = "Case5-2"
  //          y = i
  //          fmt.Printf("%s\t%s\t%d\t%s\n",line3,x,y,newMap1[y])
		  }
		  flag = 1            
		} else if newMap2[i] <= realBGst && newMap3[i] >= realBGend {
		  flag3[i] = 1
		  flag = 1
		  flag1 = 2 
  //        x = "Case6"
  //        y = i
  //        fmt.Printf("%s\t%s\t%d\t%s\n",line3,x,y,newMap1[y])         
		} 
	  } 
	}    
	if flag == 0 {
	  fmt.Printf("%s\n",line3)
	} else if flag == 1 && flag1 == 1 {
	  fmt.Printf("%s",output)
	} else if flag == 1 && flag1 == 0 {
	  output += strconv.Itoa(source_end) + "\t" + ann + "\t" + bgHomSp + "\n"
	  fmt.Printf("%s",output)
	}
	flag = 0
	flag1 = 0
	output = "" 
	realBGst = realBGend + 1
  }

  if readErr3 == io.EOF {
	break
  }
}
}
