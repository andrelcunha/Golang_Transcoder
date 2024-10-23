package converter

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"time"
)

type VideoConverter struct{}

func NewVideoConverter() *VideoConverter {
	return &VideoConverter{}
}

type VideoTask struct {
	VideoID int    `json:"video_id"`
	Path    string `json:"path"`
}

func (vc *VideoConverter) Handle(msg []byte) error {
	var task VideoTask
	err := json.Unmarshal(msg, &task)
	if err != nil {
		vc.logError(task, "Error unmarshalling task", err)
		return err
	}

	err = vc.processVideo(&task)
	if err != nil {
		vc.logError(task, "failed to process video", err)
		return err
	}
	return nil
}

func (vc *VideoConverter) processVideo(task *VideoTask) error {
	mergedFile := filepath.Join(task.Path, "merged.mp4")
	mpegDashPath := filepath.Join(task.Path, "mpegdash")

	// Merge chunks into a single file
	slog.Info("Merging chunks", slog.String("path", task.Path))
	err := vc.mergeChunks(task.Path, mergedFile)
	if err != nil {
		vc.logError(*task, "failed to merge chunks", err)
		return err
	}

	// Create mpeg-dash directory
	slog.Info("Creating mpeg-dash directory", slog.String("path", mpegDashPath))
	err = os.MkdirAll(mpegDashPath, os.ModePerm)
	if err != nil {
		vc.logError(*task, "failed to create mpeg-dash directory", err)
		return err
	}

	// Converting video to mpeg-dash
	slog.Info("Converting video to mpeg-dash", slog.String("path", mpegDashPath))
	ffmpegCmd := exec.Command(
		"ffmpeg", "-i", mergedFile,
		"-c:v", "libx264", "-c:a", "aac",
		"-f", "dash",
		filepath.Join(mpegDashPath, "output.mpd"),
	)

	output, err := ffmpegCmd.CombinedOutput()
	if err != nil {
		vc.logError(*task, "failed to convert video to mpeg-dash, output: "+string(output), err)
		return err
	}
	slog.Info("Video converted to mpeg-dash", slog.String("path", mpegDashPath))

	// Delete merged file
	slog.Info("Deleting merged file", slog.String("path", mergedFile))
	err = os.Remove(mergedFile)

	return nil
}

func (vc *VideoConverter) logError(task VideoTask, message string, err error) {
	errorData := map[string]interface{}{
		"video_id": task.VideoID,
		"error":    message,
		"details":  err.Error(),
		"time":     time.Now(),
	}

	serializedError, _ := json.Marshal(errorData)
	slog.Error("Processing error", slog.String("error_details", string(serializedError)))

	//TODO: regiister the error in the database
}

// Pasted from main.go
func (vc *VideoConverter) extractNumber(filename string) int {
	re := regexp.MustCompile(`\d+`)
	numStr := re.FindString(filepath.Base(filename)) //string
	num, err := strconv.Atoi(numStr)
	if err != nil {
		panic(err)
	}
	return num
}

// function to get the slice of chunks
func (vc *VideoConverter) getChunks(inputDir string) ([]string, error) {
	//find all the chunks in the input directory
	chunks, err := filepath.Glob(filepath.Join(inputDir, "*.chunk"))
	if err != nil {
		return nil, err
	}
	return chunks, nil
}

// function to merge the chunks into a single file
func (vc *VideoConverter) mergeChunks(inputDir string, outputFile string) error {
	chunks, err := vc.getChunks(inputDir)
	if err != nil {
		return err
	}
	//sort the chunks in sequence by the number in their filenames
	vc.sortChunks(&chunks)

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
func (vc *VideoConverter) sortChunks(chunks *[]string) {
	sort.Slice(*chunks, func(i, j int) bool {
		return vc.extractNumber((*chunks)[i]) < vc.extractNumber((*chunks)[j])
	})
}
