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
	todoFolderName   = "gamblebaybirds"
	doneFolderName   = "gamblebaybirds-processed"
	hugoPostBasePath = "content/en-us/posts"
)

func main() {
	ctx := context.Background()

	driveService, err := drive.NewService(ctx, option.WithCredentialsFile("/Users/jamespickett/.gamblebaybirds/gcp-credentials.json"))
	if err != nil {
		panic(err)
	}

	todoDriveFolder, err := getFileByName(todoFolderName, driveService)
	if err != nil {
		panic(err)
	}

	todoDriveFiles, err := getChildren(todoDriveFolder, driveService)
	if err != nil {
		panic(err)
	}

	if len(todoDriveFiles) == 0 {
		fmt.Print("\nno new files found to process\n")
		os.Exit(0)
	}

	doneDriveFolder, err := getFileByName(doneFolderName, driveService)
	if err != nil {
		panic(err)
	}

	doneDriveSubFolder, err := createDailyDoneSubFolderIfNotExists(doneDriveFolder, driveService)
	if err != nil {
		panic(err)
	}

	newPostFolderFilePath := filepath.Join("./", "content", "en-us", "posts", date())
	if err := os.MkdirAll(newPostFolderFilePath, os.ModePerm); err != nil {
		panic(err)
	}

	entries, err := os.ReadDir(newPostFolderFilePath)
	if err != nil {
		panic(err)
	}

	var imageNames []string

	for _, entry := range entries {
		if entry.IsDir() || strings.HasSuffix(entry.Name(), ".png") {
			continue
		}
		imageNames = append(imageNames, entry.Name())
	}

	for _, file := range todoDriveFiles {
		newFileName := strings.ReplaceAll(file.Name, " ", "_")

		response, err := driveService.Files.Get(file.Id).Download()
		if err != nil {
			panic(err)
		}
		defer response.Body.Close()

		imageNames = append(imageNames, newFileName)

		imagePath := filepath.Join(newPostFolderFilePath, newFileName)
		dstFile, err := os.Create(imagePath)
		if err != nil {
			panic(err)
		}

		_, err = io.Copy(dstFile, response.Body)
		if err != nil {
			panic(err)
		}

		_, err = driveService.Files.Update(file.Id, &drive.File{Name: newFileName}).RemoveParents(todoDriveFolder.Id).AddParents(doneDriveSubFolder.Id).Do()
		if err != nil {
			panic(err)
		}
	}

	newPostMdPath := filepath.Join(newPostFolderFilePath, "index.md")
	dstFile, err := os.Create(newPostMdPath)
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(dstFile, strings.NewReader(generatePostMd(imageNames)))
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

func generatePostMd(imagePaths []string) string {
	var sb strings.Builder

	sb.WriteString("---\n")

	sb.WriteString(fmt.Sprintf("date: %s\n", time.Now().Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("featured_image: %s\n", imagePaths[0]))
	sb.WriteString("summary: =)\n")
	sb.WriteString("layout: gallery\n")

	// TODO: get more tags and add descriptions
	sb.WriteString("tags:\n")
	sb.WriteString(fmt.Sprintf("  - %s\n", "Bird"))

	sb.WriteString("---\n")

	sb.WriteString("{{< gallery-grid >}}\n")
	for _, path := range imagePaths {
		sb.WriteString(fmt.Sprintf("![%s](%s)\n", path, path))
	}
	sb.WriteString("{{< /gallery-grid >}}\n")
	return sb.String()
}
