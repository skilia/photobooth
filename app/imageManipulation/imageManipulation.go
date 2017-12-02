package imageManipulation

import (
	"mime/multipart"
	"image"
	"image/jpeg"
	"image/png"
	"image/gif"
	"github.com/nfnt/resize"
	"os"
	"errors"
)

func SaveToDisk(file *multipart.FileHeader) (string, error) {
	// Convert File into Image
	imageFile, err := convertFileToImage(file)
	if err != nil {
		return "", err
	}

	// Manipulate Image
	imageFile = resize.Thumbnail(1500, 1500, imageFile, resize.Lanczos3)

	path := "/tmp/abc/"
	os.MkdirAll(path, 0777)
	filename := path + file.Filename + ".jpg"
	out, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	return filename, jpeg.Encode(out, imageFile, nil)

}

func convertFileToImage(file *multipart.FileHeader) (image.Image, error) {
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
