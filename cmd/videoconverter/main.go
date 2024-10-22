package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)

func main() {
	mergeChunks("mediatest/media/uploads/1", "merged.mp4")
}

// functiont to extract the number from the filename
func extractNumber(filename string) int {
	re := regexp.MustCompile(`\d+`)
	numStr := re.FindString(filepath.Base(filename)) //string
	num, err := strconv.Atoi(numStr)
	if err != nil {
		panic(err)
	}
	return num
}

// function to get the slice of chunks
func getChunks(inputDir string) ([]string, error) {
	//find all the chunks in the input directory
	chunks, err := filepath.Glob(filepath.Join(inputDir, "*.chunk"))
	if err != nil {
		return nil, err
	}
	return chunks, nil
}

// function to merge the chunks into a single file
func mergeChunks(inputDir string, outputFile string) error {
	chunks, err := getChunks(inputDir)
	if err != nil {
		return err
	}
	//sort the chunks in sequence by the number in their filenames
	sortChunks(&chunks)

	output, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer output.Close()

	for _, chunk := range chunks {
		input, err := os.Open(chunk)
		if err != nil {
			return fmt.Errorf("failed to open chunk: %v", err)
		}
		_, err = output.ReadFrom(input)
		if err != nil {
			return fmt.Errorf("failed to write chunk %s into merged file: %v", chunk, err)
		}
		input.Close()
	}
	return nil
}

// function to sort the chunks in sequence by the number in their filenames
func sortChunks(chunks *[]string) {
	sort.Slice(*chunks, func(i, j int) bool {
		return extractNumber((*chunks)[i]) < extractNumber((*chunks)[j])
	})
}
