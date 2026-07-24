//orthBB.go: to arrange the joint orthologous backbone file of AG
//USAGE: ./orthBB  HOM_BACKBONE_FILE  >HOM_BACKBONE_FILE

// Source_Strain	Source_st	Source_end	Annotation 	Hom_strain


package orthBB

import (
  "fmt"
  "io"
  "bufio"
  "strings"
//  "strconv"
  "os"
)

func Main() {
  usage := "orthBB: to arrange the joint orthologous backbone file of AG.\n\n" + "./orthBB  HOM_BACKBONE_FILE  >HOM_BACKBONE_FILE\n\n"
  if len(os.Args) < 2 {
    fmt.Printf("%s",usage)
    return
  }
  homFileName := os.Args[1]

  homFile, openErr := os.Open(homFileName)
  if openErr != nil {
    fmt.Printf("Error happened in opening %s!\n%s",homFileName,usage)
    return
  }
  defer homFile.Close()
  homReader := bufio.NewReader(homFile)
  
  fmt.Printf("Source_Strain\tSource_st\tSource_end\tAnnotation\tHom_strain\n")
  //sp := 0
  for {
    line1, readErr1 := homReader.ReadString('\n')
    line1 = strings.Trim(line1,"\r\n")
    line1 = strings.Trim(line1,"\n")    
    if len(line1) > 0 {
       if strings.Contains(line1,"\tORTH\t") {
         //line1Array := strings.Split(line1,"\t")
         //a, _ := strconv.Atoi(line1Array[2])
         //b, _ := strconv.Atoi(line1Array[1])
         //sp += a - b + 1
         //fmt.Printf("%s\t%d\n",line1,sp)
        fmt.Printf("%s\n",line1)
       }
    }

    if readErr1 == io.EOF {
      break
    }
  } 
}  