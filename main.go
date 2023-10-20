package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
)

type File struct {
	Path     string
	Contents []byte
}

func readFiles(directory string) ([]File, error) {
	var files []File

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			contents, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			files = append(files, File{Path: path, Contents: contents})
		}

		return nil
	})

	return files, err
}

func calculateSimilarity(file1 File, file2 File) float64 {
	bytesFile1 := make(map[byte]uint64)
	bytesFile2 := make(map[byte]uint64)

	for _, b := range file1.Contents {
		bytesFile1[b]++
	}
	for _, b := range file2.Contents {
		bytesFile2[b]++
	}

	maxLen := Max(len(file1.Contents), len(file2.Contents))
	commonBytes := uint64(0)

	for key := range bytesFile1 {
		commonBytes += Min(bytesFile2[key], bytesFile1[key])
	}

	return (float64(commonBytes) / float64(maxLen)) * 100.0
}

func Min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

func Max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func main() {
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Usage:")
		fmt.Fprintf(flag.CommandLine.Output(), "./%s [options] \n\n", path.Base(os.Args[0]))
		fmt.Fprintln(flag.CommandLine.Output(), "Options:")
		flag.PrintDefaults()
	}

	var dir1 string
	flag.StringVar(&dir1, "dir1", dir1, "(required) path to first dir")

	var dir2 string
	flag.StringVar(&dir2, "dir2", dir2, "(required) path to second dir")

	var similarityThreshold float64
	flag.Float64Var(&similarityThreshold, "similarity", similarityThreshold, "(required) percentage of similarity for files differences")

	flag.Parse()

	if flag.NFlag() != 3 {
		flag.Usage()
		os.Exit(1)
	}

	files1, err := readFiles(dir1)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading files from directory 1:", err)
		return
	}

	files2, err := readFiles(dir2)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading files from directory 2:", err)
		return
	}

	identicalFiles := make(map[string]string)
	similarFiles := make(map[string]map[string]float64)
	similarFilesFromDir1 := make(map[string]struct{})
	similarFilesFromDir2 := make(map[string]struct{})

	for _, file1 := range files1 {
		for _, file2 := range files2 {
			hash1 := sha256.Sum256(file1.Contents)
			hash2 := sha256.Sum256(file2.Contents)

			if hash1 == hash2 {
				identicalFiles[file1.Path] = file2.Path

				similarFilesFromDir1[file1.Path] = struct{}{}
				similarFilesFromDir2[file2.Path] = struct{}{}
			} else {
				similarity := calculateSimilarity(file1, file2)
				if similarity >= similarityThreshold {
					if similarFiles[file1.Path] == nil {
						similarFiles[file1.Path] = make(map[string]float64)
					}
					similarFiles[file1.Path][file2.Path] = similarity

					similarFilesFromDir1[file1.Path] = struct{}{}
					similarFilesFromDir2[file2.Path] = struct{}{}
				}
			}
		}
	}

	// Identical files
	countPrints := 0
	fmt.Println("Identical files:")
	for file1, file2 := range identicalFiles {
		fmt.Printf("%s - %s\n", file1, file2)
		countPrints++
	}
	if countPrints == 0 {
		fmt.Println("have 0 files")
	}

	// Similar files
	countPrints = 0
	fmt.Println("\nSimilar files:")
	for file1, similarFiles := range similarFiles {
		for file2, similarity := range similarFiles {
			fmt.Printf("%s - %s - %.2f%% similarity\n", file1, file2, similarity)
			countPrints++
		}
	}
	if countPrints == 0 {
		fmt.Println("have 0 files")
	}

	// Files only in dir1
	countPrints = 0
	fmt.Println("\nFiles present in", dir1, "but not in", dir2, ":")
	for _, file1 := range files1 {
		if _, exists := similarFilesFromDir1[file1.Path]; !exists {
			fmt.Println(file1.Path)
			countPrints++
		}
	}
	if countPrints == 0 {
		fmt.Println("have 0 files")
	}

	// Files only in dir2
	countPrints = 0
	fmt.Println("\nFiles present in", dir2, "but not in", dir1, ":")
	for _, file2 := range files2 {
		if _, exists := similarFilesFromDir2[file2.Path]; !exists {
			fmt.Println(file2.Path)
			countPrints++
		}
	}
	if countPrints == 0 {
		fmt.Println("have 0 files")
	}
}
