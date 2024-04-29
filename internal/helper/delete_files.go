package helper

import (
	"fmt"
	"os"
)

func DeleteFiles(dir string)  {
    // Open the directory
    d, err := os.Open(dir)
    if err != nil {
        fmt.Println("Error opening directory:", err)
        return
    }
    defer d.Close()

    // Read all files in the directory
    files, err := d.Readdir(-1)
    if err != nil {
        fmt.Println("Error reading directory contents:", err)
        return
    }

    // Remove each file
    for _, file := range files {
        err := os.Remove(dir + "/" + file.Name())
        if err != nil {
            fmt.Println("Error removing file:", err)
            return
        }
        fmt.Println("Removed file:", file.Name())
    }
}