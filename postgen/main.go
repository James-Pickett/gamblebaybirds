package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

const (
	todoFolderName, doneFolderName = "gamblebaybirds", "gamblebaybirds-processed"
)

func main() {
	ctx := context.Background()
	driveService, err := drive.NewService(ctx, option.WithCredentialsFile("/Users/jamespickett/.gamblebaybirds/gcp-credentials.json"))
	if err != nil {
		panic(err)
	}

	todoFolder, err := getFileByName(todoFolderName, driveService)
	if err != nil {
		panic(err)
	}

	todoFiles, err := getChildren(todoFolder, driveService)
	if err != nil {
		panic(err)
	}

	if len(todoFiles) == 0 {
		fmt.Print("\nno new files found to process")
		os.Exit(1)
	}

	doneFolder, err := getFileByName(doneFolderName, driveService)
	if err != nil {
		panic(err)
	}

	doneSubFodler, err := createDailyDoneSubFolderIfNotExists(doneFolder, driveService)
	if err != nil {
		panic(err)
	}

	newPostFolderPath := filepath.Join(rootDir(), "content", "moments", date())
	if err := os.MkdirAll(newPostFolderPath, os.ModePerm); err != nil {
		panic(err)
	}

	newPostImagesFolderPath := filepath.Join(newPostFolderPath, "images")
	if err := os.MkdirAll(newPostImagesFolderPath, os.ModePerm); err != nil {
		panic(err)
	}

	entries, err := os.ReadDir(newPostImagesFolderPath)
	if err != nil {
		panic(err)
	}

	var imagePaths []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		imagePaths = append(imagePaths, filepath.Join("moments", date(), "images", entry.Name()))
	}

	for _, file := range todoFiles {
		newFileName := strings.ReplaceAll(file.Name, " ", "_")

		response, err := driveService.Files.Get(file.Id).Download()
		if err != nil {
			panic(err)
		}
		defer response.Body.Close()

		imagePath := filepath.Join(newPostImagesFolderPath, newFileName)
		imagePaths = append(imagePaths, filepath.Join("moments", date(), "images", newFileName))

		dstFile, err := os.Create(imagePath)
		if err != nil {
			panic(err)
		}

		_, err = io.Copy(dstFile, response.Body)
		if err != nil {
			panic(err)
		}

		_, err = driveService.Files.Update(file.Id, &drive.File{Name: newFileName}).RemoveParents(todoFolder.Id).AddParents(doneSubFodler.Id).Do()
		if err != nil {
			panic(err)
		}
	}

	newPostMdPath := filepath.Join(newPostFolderPath, fmt.Sprintf("%s.md", date()))
	dstFile, err := os.Create(newPostMdPath)
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(dstFile, strings.NewReader(generatePostMd(imagePaths)))
	if err != nil {
		panic(err)
	}
}

func createDailyDoneSubFolderIfNotExists(parent *drive.File, driveService *drive.Service) (*drive.File, error) {
	// test to see if file already exists
	file, err := getFileByName(date(), driveService, parent.Id)
	if err == nil {
		return file, nil
	}

	file, err = driveService.Files.Create(&drive.File{
		MimeType: "application/vnd.google-apps.folder",
		Name:     date(),
		Parents:  []string{parent.Id},
	}).SupportsAllDrives(true).Do()

	if err != nil {
		return nil, fmt.Errorf("creating folder %s: %w", date(), err)
	}

	return file, nil
}

func getChildren(parent *drive.File, driveService *drive.Service) ([]*drive.File, error) {
	result, err := driveService.Files.List().Q(fmt.Sprintf("'%s' in parents", parent.Id)).Do()
	if err != nil {
		return nil, err
	}

	return result.Files, nil
}

func getFileByName(name string, service *drive.Service, parents ...string) (*drive.File, error) {
	queryString := fmt.Sprintf("name = '%s'", name)

	for _, parent := range parents {
		queryString = fmt.Sprintf("%s and '%s' in parents", queryString, parent)
	}

	list, err := service.Files.List().Q(queryString).Do()

	if err != nil {
		return nil, err
	}

	if len(list.Files) == 0 {
		return nil, fmt.Errorf("no file found with name %s", name)
	}

	if len(list.Files) > 1 {
		return nil, fmt.Errorf("more than one file found with name '%s'", name)
	}

	return list.Files[0], nil
}

func date() string {
	return time.Now().Format("2006-01-02")
}

func rootDir() string {
	path, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	if filepath.Base(path) == "postgen" {
		return "../"
	}

	return "./"
}

func generatePostMd(imagePaths []string) string {
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("name: %s\n", date()))
	sb.WriteString(fmt.Sprintf("date: %s\n", time.Now().Format(time.RFC3339)))
	sb.WriteString("pictures:\n")

	for _, imagePath := range imagePaths {
		sb.WriteString(fmt.Sprintf("  - %s\n", imagePath))
	}
	sb.WriteString("---\n")
	return sb.String()
}
