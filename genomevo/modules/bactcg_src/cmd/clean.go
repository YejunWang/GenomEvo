package cmd

import (
    "bufio"
    "fmt"
    "os"
    "path/filepath"
    "strings"

    "github.com/spf13/cobra"
)

var inputDir string

var cleanCmd = &cobra.Command{
    Use:   "clean",
    Short: "Clean whitespace and format headers in FASTA files",
    Long:  `Equivalent to del.py and remove_empty_lines.py: Removes all empty lines and trims sequences header at first space.`,
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("Cleaning files in", inputDir)
        err := cleanDir(inputDir)
        if err != nil {
            fmt.Println("Error:", err)
        } else {
            fmt.Println("Cleaning complete.")
        }
    },
}

func init() {
    rootCmd.AddCommand(cleanCmd)
    cleanCmd.Flags().StringVarP(&inputDir, "input", "i", "dataset/1", "Directory to clean")
}

func cleanDir(dir string) error {
    files, err := os.ReadDir(dir)
    if err != nil {
        return err
    }
    
    for _, f := range files {
        if !f.IsDir() {
            path := filepath.Join(dir, f.Name())
            if err := processFasta(path); err != nil {
                fmt.Printf("Error processing %s: %v\n", f.Name(), err)
            }
        }
    }
    return nil
}

func processFasta(path string) error {
    file, err := os.Open(path)
    if err != nil {
        return err
    }
    defer file.Close()
    
    var lines []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        text := scanner.Text()
        if strings.TrimSpace(text) == "" {
            continue
        }
        if strings.HasPrefix(text, ">") {
            parts := strings.SplitN(text, " ", 2)
            lines = append(lines, parts[0])
        } else {
            lines = append(lines, text)
        }
    }
    
    if err := scanner.Err(); err != nil {
        return err
    }
    
    out, err := os.Create(path)
    if err != nil {
        return err
    }
    defer out.Close()
    
    writer := bufio.NewWriter(out)
    for _, l := range lines {
        writer.WriteString(l + "\n")
    }
    writer.Flush()
    
    fmt.Println("Successfully cleaned", path)
    return nil
}
