package controllers

import (
	"github.com/revel/revel"
	"os"
	"mime/multipart"
)

type Multiple struct {
	App
}

func (c *Multiple) Upload() revel.Result {
	return c.Render()
}

func (c *Multiple) HandleUpload() revel.Result {
	var files [][]byte
	c.Params.Bind(&files, "file")

	// Make sure at least 2 but no more than 3 files are submitted.
	c.Validation.MinSize(files, 2).Message("You cannot submit less than 2 files")
	//c.Validation.MaxSize(files, 3).Message("You cannot submit more than 3 files")

	// Handle errors.
	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect((*Multiple).Upload)
	}

	// Prepare result.
	filesInfo := make([]FileInfo, len(files))
	for i, _ := range files {
		aFile := c.Params.Files["file[]"][i]
		a, err := writeFile(aFile)
		if (err != nil) {
			filesInfo[i] = FileInfo{
				ContentType: aFile.Header.Get("Content-Type"),
				Filename:    err.Error(),
				Size:        len(files[i]),
			}
			continue;
		}

		filesInfo[i] = FileInfo{
			ContentType: aFile.Header.Get("Content-Type"),
			Filename:    a,
			Size:        len(files[i]),
		}
	}

	return c.RenderJSON(map[string]interface{}{
		"Count":  len(files),
		"Files":  filesInfo,
		"Status": "Successfully uploaded",
	})
}

func writeFile(file *multipart.FileHeader) (string, error) {

	sourceFile, err := file.Open()
	targetFile, err := os.Create("/tmp/abc/" + file.Filename)
	if (err != nil) {
		return "", err
	}

	defer targetFile.Close()
	// TODO: For full file load, use file.Size as buffer size
	buffer := make([]byte, 4096)
	for {
		bytesRead, err := sourceFile.Read(buffer)

		if err != nil {
			if err.Error() == "EOF" {
				break
			}

			return "", err
		}

		targetFile.Write(buffer[0:bytesRead])
	}

	targetFile.Sync()

	return targetFile.Name(), nil;
}

type FileInfo struct {
	ContentType string
	Filename    string
	RealFormat  string `json:",omitempty"`
	Resolution  string `json:",omitempty"`
	Size        int
	Status      string `json:",omitempty"`
}
