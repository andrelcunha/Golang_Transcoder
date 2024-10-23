package converter

import (
	"encoding/json"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/andrelcunha/Golang_Transcoder/internal/merger"
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
	vm := merger.NewVideoMerger()

	mergedFile := filepath.Join(task.Path, "merged.mp4")
	mpegDashPath := filepath.Join(task.Path, "mpegdash")

	// Merge chunks into a single file
	slog.Info("Merging chunks", slog.String("path", task.Path))
	err := vm.MergeChunks(task.Path, mergedFile)
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
	if err != nil {
		vc.logError(*task, "failed to delete merged file", err)
		return err
	}
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
