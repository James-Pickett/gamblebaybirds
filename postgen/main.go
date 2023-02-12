package main

import (
	"context"
	"fmt"
	"os"
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

	for _, file := range todoFiles {
		_, err := driveService.Files.Update(file.Id, &drive.File{}).RemoveParents(todoFolder.Id).AddParents(doneSubFodler.Id).Do()
		if err != nil {
			panic(err)
		}
	}
}

func createDailyDoneSubFolderIfNotExists(parent *drive.File, driveService *drive.Service) (*drive.File, error) {
	fileName := time.Now().Format("2006-01-02")

	// test to see if file already exists
	file, err := getFileByName(fileName, driveService, parent.Id)
	if err == nil {
		return file, nil
	}

	file, err = driveService.Files.Create(&drive.File{
		MimeType: "application/vnd.google-apps.folder",
		Name:     fileName,
		Parents:  []string{parent.Id},
	}).SupportsAllDrives(true).Do()

	if err != nil {
		return nil, fmt.Errorf("creating folder %s: %w", fileName, err)
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
