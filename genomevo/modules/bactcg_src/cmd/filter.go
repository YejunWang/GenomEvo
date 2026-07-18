package cmd

import (
    "fmt"
    "os"
    "path/filepath"
    "github.com/spf13/cobra"
)

var filterDir string

var filterCmd = &cobra.Command{
    Use:   "filter",
    Short: "Filter out protein files smaller than 90% of the average file size",
    Long:  `Port of 0-1.filterate_file_M1.py logic. Reads all files in the directory, calculates average size, and deletes those smaller than 90% of average.`,
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("Filtering files in", filterDir)
        err := FilterFiles(filterDir)
        if err != nil {
            fmt.Println("Error:", err)
        }
    },
}

func init() {
    rootCmd.AddCommand(filterCmd)
    filterCmd.Flags().StringVarP(&filterDir, "input", "i", "input/seq/prot_file/", "Directory to filter")
}

func FilterFiles(dir string) error {
    files, err := os.ReadDir(dir)
    if err != nil {
        return err
    }
    
    var totalSize int64
    count := 0
    
    for _, f := range files {
        if !f.IsDir() {
            info, err := f.Info()
            if err == nil {
                totalSize += info.Size()
                count++
            }
        }
    }
    
    if count == 0 {
        return fmt.Errorf("no files found in %s", dir)
    }
    
    avgSize := float64(totalSize) / float64(count)
    cutoff := avgSize * 0.9
    
    fmt.Printf("Average size: %.2f bytes. Cutoff: %.2f bytes.\n", avgSize, cutoff)
    
    removeLog, err := os.Create("remove_id.txt")
    if err != nil {
        return err
    }
    defer removeLog.Close()
    
    for _, f := range files {
        if !f.IsDir() {
            info, err := f.Info()
            if err == nil && float64(info.Size()) < cutoff {
                path := filepath.Join(dir, f.Name())
                fmt.Printf("Removing %s (size: %d)\n", f.Name(), info.Size())
                
                // Write ID to log (strip extension)
                name := f.Name()
                ext := filepath.Ext(name)
                id := name[0 : len(name)-len(ext)]
                removeLog.WriteString(id + "\n")
                
                os.Remove(path)
            }
        }
    }
    return nil
}
