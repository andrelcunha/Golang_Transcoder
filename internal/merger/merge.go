package merger

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)

type VideoMerger struct{}

func NewVideoMerger() *VideoMerger {
	return &VideoMerger{}
}

// function to merge the chunks into a single file
func (vm *VideoMerger) MergeChunks(inputDir string, outputFile string) error {
	chunks, err := vm.getChunks(inputDir)
	if err != nil {
		return err
	}
	//sort the chunks in sequence by the number in their filenames
	vm.sortChunks(&chunks)

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

func (vm *VideoMerger) extractNumber(filename string) int {
	re := regexp.MustCompile(`\d+`)
	numStr := re.FindString(filepath.Base(filename)) //string
	num, err := strconv.Atoi(numStr)
	if err != nil {
		panic(err)
	}
	return num
}

// function to get the slice of chunks
func (vm *VideoMerger) getChunks(inputDir string) ([]string, error) {
	//find all the chunks in the input directory
	chunks, err := filepath.Glob(filepath.Join(inputDir, "*.chunk"))
	if err != nil {
		return nil, err
	}
	return chunks, nil
}

// function to sort the chunks in sequence by the number in their filenames
func (vm *VideoMerger) sortChunks(chunks *[]string) {
	sort.Slice(*chunks, func(i, j int) bool {
		return vm.extractNumber((*chunks)[i]) < vm.extractNumber((*chunks)[j])
	})
}
