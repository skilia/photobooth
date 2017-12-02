package controllers

import (
	"github.com/revel/revel"
	//	"github.com/skilia/photobooth/app/printer"
	"github.com/skilia/photobooth/app/imageManipulation"
	"mime/multipart"
	"os"
	"fmt"
)

type Multiple struct {
	*revel.Controller
}

type FileInfo struct {
	ContentType string
	Filename    string
	RealFormat  string `json:",omitempty"`
	Resolution  string `json:",omitempty"`
	Size        int64
	Status      string `json:",omitempty"`
}

type Response struct {
	Files  []FileInfo `json:",omitempty"`
	Errors []string   `json:",omitempty"`
	Count  int
	Status string
}

func (c *Multiple) Upload() revel.Result {
	return c.Render()
}

func (c *Multiple) HandleUpload() revel.Result {
	response := Response{}
	// Prepare result.
	for _, fileHeaders := range c.Params.Files {
		for _, aFile := range fileHeaders {
			response.Count++

			aFileInfo, err := handleSingleFile(aFile)
			if err != nil {
				response.Errors = append(response.Errors, err.Error())
				continue
			}

			response.Files = append(response.Files, aFileInfo)
		}
	}

	if len(response.Errors) < 1 {
		response.Status = "Successfully uploaded"
	} else {
		response.Status = "Errors occurred"
	}

	return c.RenderJSON(response)
}

func handleSingleFile(aFile *multipart.FileHeader) (FileInfo, error) {
	// Convert image and save to disk
	newFilePath, err := imageManipulation.SaveToDisk(aFile)
	if err != nil {
		return FileInfo{}, err
	}

	//printer.AddImage(newFilePath)

	fileStruct := FileInfo{
		ContentType: aFile.Header.Get("Content-Type"),
		Filename:    newFilePath,
		Size: getFileSize(aFile),
	}

	return fileStruct, nil
}
func getFileSize(fileHeader *multipart.FileHeader) int64 {
	mfile, _ := fileHeader.Open() //fh *multipart.FileHeader
	switch t := mfile.(type) {
	case *os.File:
		fi, _ := t.Stat()
		return fi.Size()
	default:
		sr, _ := mfile.Seek(0,0)
		fmt.Println(sr)
		revel.AppLog.Error("Unable to get file size", "t", t, "sr", sr)
		return 0
	}
}
