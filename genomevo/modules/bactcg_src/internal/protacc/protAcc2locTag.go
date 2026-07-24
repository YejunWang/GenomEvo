/* protAcc2locTag.go
   The program was written to re-annotate the protein names of proteome fasta file from accession numbers (WP_) to unique locus tags.

   USAGE:	./protAcc2locTag	 ACC_LOCUSTAG_FILE	RAW_PROTEOME_FASTA_FILE		[OPTIONS: chr | OTHERS]		[OPTIONS: nr]

   One can use the program without the 4th argument and no replicon will be selected; alternatively, 'chr' or OTHERS could be specified to selectively re-annotate the proteins present in the desired replicon and output to a new file.
   The program could also be run without the 5th argument, so the proteins in different loci but with the same sequence or protein accession will not be filtered but output with a '|' separated locus tags in the annotation line. If the 'nr' is specified, the repeat proteins will be filtered. 
 
   ACC_LOCUSTAG_FILE is generated and downloaded from NCBI Genome database.
*/

package protacc

import (
  "fmt"
  "os"
  "io"
  "bufio"
  "strings"
)

func RunProtAcc () {
  usage := "\nprotAcc2locTag.go: Re-annotating protein names from accessions to locus tags. \n\n" + "USAGE:\nprotAcc2locTag  ACC_LOCUSTAG_FILE  RAW_PROTEOME_FASTA_FILE  [OPTIONS: chr | OTHERS]  [OPTIONS: nr]\n\n"
  var arg int
  var replicon string
  if len(os.Args) < 3 || len(os.Args) > 5 {
  	fmt.Printf("%s",usage)
  	return
  }
  if len(os.Args) == 5 {
  	replicon = os.Args[3]
  	if strings.EqualFold(os.Args[4],"nr") {
      arg = 5
  	} else {
  	  fmt.Printf("%s",usage)
  	  return	
  	}
  } else if len(os.Args) == 4 {
    if strings.EqualFold(os.Args[3],"nr") {
      arg = 41
  	} else {
  	  replicon = os.Args[3]
  	  arg = 42	
  	}
  } else {
    arg = 3
  }

  acc2loctagFile := os.Args[1]
  rawProtFile := os.Args[2]

  acc2loctag, openErr1 := os.Open(acc2loctagFile)
  if openErr1 != nil {
    fmt.Printf("Error happened in opening the file %s!\n",acc2loctagFile)
    fmt.Printf("%s",usage)
    return
  }
  defer acc2loctag.Close()

  buf_acc2loctag := bufio.NewReader(acc2loctag)
  acc2loctag_map := make(map[string]string)
  replicon_map := make(map[string]string)
  repeatProt := make(map[string]int) 
  for {
  	tabLine, readErr1 := buf_acc2loctag.ReadString('\n')
  	tabLine = strings.Trim(tabLine,"\r\n")
  	tabLine = strings.Trim(tabLine,"\n")
    if len(tabLine) > 0 {
      tabLineArray := strings.Split(tabLine,"\t")
      if len(tabLineArray) >= 10 && !strings.Contains(tabLineArray[0],"#") {
        tab_replicon := tabLineArray[0]
        tab_loctag := tabLineArray[7]
        tab_protAcc := tabLineArray[8]
        if len(acc2loctag_map[tab_protAcc]) > 0 {
          acc2loctag_map[tab_protAcc] += "|" + tab_loctag
          replicon_map[tab_protAcc] += "|" + tab_replicon
          repeatProt[tab_protAcc] = 1
        } else {
          acc2loctag_map[tab_protAcc] = tab_loctag
          replicon_map[tab_protAcc] = tab_replicon
        }       
      }
    } 
  	if readErr1 == io.EOF {
      break
  	}
  }

  rawProt, openErr2 := os.Open(rawProtFile)
  if openErr2 != nil {
    fmt.Printf("Error happened in opening the file %s!\n",rawProtFile)
    fmt.Printf("%s",usage)
    return 	
  }
  defer rawProt.Close()
  
  buf_rawProt := bufio.NewReader(rawProt)
  flag := 0
  for {
  	protLine, readErr2 := buf_rawProt.ReadString('\n')
  	protLine = strings.Trim(protLine,"\r\n")
  	protLine = strings.Trim(protLine,"\n")
    if strings.Contains(protLine,">") {
  	  protLine = strings.Replace(protLine,">","",-1)
  	  if len(protLine) > 0 {
  	  	annArray := strings.Fields(protLine)
  	  	if len(acc2loctag_map[annArray[0]]) > 0 {
          switch arg {
            case 3:	
              flag = 1
              fmt.Printf(">%s\n",acc2loctag_map[annArray[0]])
            case 5: 
              if repeatProt[annArray[0]] == 0 && strings.EqualFold(replicon_map[annArray[0]],replicon) {
              	flag = 1
              	fmt.Printf(">%s\n",acc2loctag_map[annArray[0]])
              } else {
              	flag = 0
              } 
            case 41:
              if repeatProt[annArray[0]] == 0 {
              	flag = 1
              	fmt.Printf(">%s\n",acc2loctag_map[annArray[0]])
              }	else {
              	flag = 0
              }
            case 42:
              if strings.Contains(replicon_map[annArray[0]],replicon) {
              	flag = 1
              	fmt.Printf(">%s\n",acc2loctag_map[annArray[0]])
              } else {
              	flag = 0
              }
          }
  	  	}
  	  } else {
  	  	flag = 0
  	  }     	
    } else {
      if flag == 1 {
      	fmt.Printf("%s\n",protLine)
      }
    }

  	if readErr2 == io.EOF {
      break
  	}
  }
}

