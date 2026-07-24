//revcomp	GENOME_FASTA	>GENOME_RC_FASTA

package revcomp

import (
  "fmt"
  "os"
  "io"
  "bufio"
  "strings"
)

func Main() {
  if len(os.Args) <= 1 {
    fmt.Printf("Input argument not right!\n")
    return
  }
  inputFileName := os.Args[1]
  inputFile, openErr := os.Open(inputFileName)

  if openErr != nil {
    fmt.Printf("Error happened in opening %s!\n",inputFileName)
    return
  }
  defer inputFile.Close()

  inputReader := bufio.NewReader(inputFile)
  var seq,ann string
  for {
    line, readErr := inputReader.ReadString('\n')
    line = strings.Trim(line,"\r\n")
    line = strings.Trim(line,"\n")
    if (strings.Contains(line,">")) {
      if len(seq) > 0 {
        seq = reverse(seq)
        seq = comp(seq)
        seq = substrPrint(seq,70)
        fmt.Printf("%s\n",ann)
        fmt.Printf("%s\n",seq)
      }
      ann = line
      seq = ""
    } else {
      seq += line
    }  
    if readErr == io.EOF {
      if len(seq) > 0 {
        seq = reverse(seq)
        seq = comp(seq)
        seq = substrPrint(seq,70)
        fmt.Printf("%s\n",ann)
        fmt.Printf("%s\n",seq)
      }  
      break
    }
  }
}

func reverse (s string) string {
	r := []rune(s)
	for from,to := 0,len(r)-1; from < to; from,to = from+1,to-1 {
	   r[from],r[to] = r[to],r[from]
	}
	return string(r)
}

func comp (s string) string {
	s = strings.Replace(s,"A","X",-1)
	s = strings.Replace(s,"T","A",-1)
	s = strings.Replace(s,"X","T",-1)
    s = strings.Replace(s,"C","X",-1)
	s = strings.Replace(s,"G","C",-1)
	s = strings.Replace(s,"X","G",-1)
	s = strings.Replace(s,"a","x",-1)
	s = strings.Replace(s,"t","a",-1)
	s = strings.Replace(s,"x","t",-1)
    s = strings.Replace(s,"c","x",-1)
	s = strings.Replace(s,"g","c",-1)
	s = strings.Replace(s,"x","g",-1)
	return s	
}

func substrPrint (s string, l int) string {
	var s1 string
	line := (len(s) - len(s) % l) / l
    for i := 0; i < line; i++ {
       s1 = s1 + s[i*l:(i+1)*l] + "\n"
    }
    if len(s)%l != 0 {
    	s1 = s1 + s[line*l:] + "\n"
    }
    return s1
}