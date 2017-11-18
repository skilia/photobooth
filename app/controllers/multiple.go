package controllers

import (
	"github.com/revel/revel"
	"os"
	"mime/multipart"
	"image"
	"image/jpeg"
	"image/png"
	"image/gif"
	"errors"
	"github.com/nfnt/resize"
	"github.com/skilia/photobooth/app/printer"
)

type Multiple struct {
	*revel.Controller
}

func (c *Multiple) Upload() revel.Result {
	return c.Render()
}

func (c *Multiple) HandleUpload() revel.Result {
	var files [][]byte
	c.Params.Bind(&files, "file")

	// Make sure at least 2 but no more than 3 files are submitted.
	//c.Validation.MinSize(files, 2).Message("You cannot submit less than 2 files")
	//c.Validation.MaxSize(files, 3).Message("You cannot submit more than 3 files")

	// Handle errors.
	//if c.Validation.HasErrors() {
	//	c.Validation.Keep()
	//	c.FlashParams()
	//	return c.Redirect((*Multiple).Upload)
	//}

	// Prepare result.
	filesInfo := make([]FileInfo, len(files))
	for i, _ := range files {
		aFile := c.Params.Files["file[]"][i]
		newFilePath, err := trans(aFile)
		if (err != nil) {
			filesInfo[i] = FileInfo{
				ContentType: aFile.Header.Get("Content-Type"),
				Filename:    err.Error(),
				Size:        len(files[i]),
			}

			continue
		}

		printer.AddImage(newFilePath)

		filesInfo[i] = FileInfo{
			ContentType: aFile.Header.Get("Content-Type"),
			Filename:    newFilePath,
			Size:        len(files[i]),
		}
	}

	return c.RenderJSON(map[string]interface{}{
		"Count":  len(files),
		"Files":  filesInfo,
		"Status": "Successfully uploaded",
	})
}

func fileToImage(file *multipart.FileHeader) (image.Image, error) {
	f, err := file.Open()
	if err != nil {
		return nil, err
	}

	mimeType := file.Header.Get("Content-Type")

	var decodedImage image.Image

	switch mimeType {
	case "image/jpeg":
		decodedImage, err = jpeg.Decode(f)
	case "image/png":
		decodedImage, err = png.Decode(f)
	case "image/gif":
		decodedImage, err = gif.Decode(f)
	default:
		return nil, errors.New("Unsupported MIME Type: '" + mimeType + "'")
	}

	return decodedImage, err
}

func trans(file *multipart.FileHeader) (string, error) {
	imageFile, err := fileToImage(file)
	if err != nil {
		return "", err
	}

	imageFile = resize.Thumbnail(500, 500, imageFile, resize.Lanczos3)

	filename := "/tmp/abc/" + file.Filename + ".jpg"
	out, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	return filename, jpeg.Encode(out, imageFile, nil)

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

