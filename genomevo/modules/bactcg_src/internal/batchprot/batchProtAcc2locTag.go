//batchProtAcc2locTag.go: batch execution of protAcc2locTag program.
//USAGE:  ./batchProtAcc2locTag

/* The protAcc2locTag program should be transferred into the current directory, and there are also three sub-folders, 
   namely 'prot_seq', 'prot_list' and 'result'. The proteome sequenes (FASTA format) are recorded in files stored in 
   'prot_seq' while the protein coordinates (and Accession-LocusTag maps) are recorded in files stored in 'prot_list'. 
   The 'result' is used for storing resulted protein files. For each strain, both the sequence and the list file should 
   be named with the strain name as prefix, followed by '.protein.fasta' and '.prot.table.txt' respectively, while 
   'result' file is with '.faa' suffix after the strain name.

   After running the program, a new sub-folder named 'output' will be generated, where the re-annoated sequence files 
   were stored.
*/

package batchprot
	
import (
  "fmt"
  "io/ioutil"
  "os"
  //"path/filepath"
  "strings"
  "os/exec"
  "log"
)

func RunBatchProt () {
  
  listFiles, err1 := ListDir("./prot_list",".prot.table.txt")
  if err1 != nil {
    fmt.Printf("Error happened in the file listing step!\n")
  	return
  }
  
  var filePrefix, seqFile, resultFile string
  cmd := exec.Command("cd",".")
  for _,listFile := range listFiles {
  	if len(listFile) > 0 {
      filePrefix = strings.Replace(listFile,".prot.table.txt","",-1)
      filePrefix = strings.Replace(filePrefix,"./prot_list/","",-1)
      seqFile = "./prot_seq/" + filePrefix + ".protein.fasta"
      resultFile = "./result/" + filePrefix + ".faa"
      cmd = exec.Command("./protAcc2locTag", listFile, seqFile, "chr", "nr")
      buf, err2 := cmd.Output()
      if err2 != nil {
        log.Fatal(err2)
      } else {
        writeFile(resultFile,string(buf))
      }
    }  
  }
}   

func ListDir(dirPth string, suffix string) (files []string, err error) {
  files = make([]string,0, 1000)
  dir, err := ioutil.ReadDir(dirPth)
  if err != nil {
  	return nil, err
  }
  PthSep := string(os.PathSeparator)
  suffix = strings.ToUpper(suffix)

  for _,fi := range dir {
  	if fi.IsDir() {
  	  continue
  	}
  	if strings.HasSuffix(strings.ToUpper(fi.Name()),suffix) {
  	  files = append(files,dirPth+PthSep+fi.Name())
    }
  }
  return files,nil
}

func writeFile(newFile string, seq string) {
  outFile, err := os.Create(newFile)
  defer outFile.Close()
  if err != nil {
  	fmt.Printf("Errors happened in creating result files!\n")
  }
  outFile.WriteString(seq)
}
