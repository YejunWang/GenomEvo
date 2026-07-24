//homBB.go: to arrange the joint homologous backbone file of AG
//USAGE: ./homoBB  JOINT_BACKBONE_TXT_FILE  Source_STRAIN_NAME >HOM_BACKBONE_FILE

// Source_Strain	Source_st	Source_end	Annotation 	Hom_strain


package homBB

import (
  "fmt"
  "io"
  "bufio"
  "strings"
  "os"
)

func Main() {
  usage := "homBB: to arrange the joint homologous backbone file of AG.\n\n" + "./homoBB  JOINT_BACKBONE_TXT_FILE  Source_STRAIN_NAME >HOM_BACKBONE_FILE\n\n"
  if len(os.Args) < 3 {
    fmt.Printf("%s",usage)
    return
  }
  homFileName := os.Args[1]
  sourceStr := os.Args[2]

  homFile, openErr := os.Open(homFileName)
  if openErr != nil {
    fmt.Printf("Error happened in opening %s!\n%s",homFileName,usage)
    return
  }
  defer homFile.Close()
  homReader := bufio.NewReader(homFile)
  
  fmt.Printf("Source_Strain\tSource_st\tSource_end\tAnnotation\tHom_strain\n")
  for {
    line1, readErr1 := homReader.ReadString('\n')
    line1 = strings.Trim(line1,"\r\n")
    line1 = strings.Trim(line1,"\n")    
    if len(line1) > 0 {
      // if strings.Contains(line1,"\tORTH\t0") {
      fmt.Printf("%s\t%s\n",sourceStr,line1)
      //}
    }

    if readErr1 == io.EOF {
      break
    }
  } 
}  