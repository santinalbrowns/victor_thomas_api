package helper

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"time"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/nfnt/resize"
)

func OpenImage(filename, fileType string) (image.Image, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	switch fileType {
	case "image/jpeg":
		return jpeg.Decode(file)
	case "image/png":
		return png.Decode(file)
	default:
		return nil, fmt.Errorf("unsupported file type")
	}
}

func CreateImageVersions(src image.Image, filename, ext string, path string) error {
	qualities := []int{ /* 100, 75, 50, */ 25}

	for _, quality := range qualities {
		resized := resize.Resize(uint(src.Bounds().Dx()*quality/100), 0, src, resize.Lanczos3)
		outputPath := filepath.Join(path, filename)
		err := SaveImage(resized, outputPath, ext)
		if err != nil {
			return err
		}
	}
	return nil
}

func SaveImage(img image.Image, outputPath, ext string) error {
	switch ext {
	case ".jpg":
		return imaging.Save(img, outputPath, imaging.JPEGQuality(80))
	case ".png":
		return imaging.Save(img, outputPath, imaging.PNGCompressionLevel(png.BestCompression))
	default:
		return fmt.Errorf("unsupported file type")
	}
}

func UploadImage(file multipart.File, header textproto.MIMEHeader) (*string, error) {
	defer file.Close()

	// Determine the file type
	fileType := header.Get("Content-Type")
	var ext string
	switch fileType {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	default:
		return nil, fmt.Errorf("unsupported file type")
	}

	// Generate a unique file name using UUID and timestamp
	uuid := uuid.New()
	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("%s_%s%s", timestamp, uuid.String(), ext)

	uploadPath := os.Getenv("UPLOADS_PATH")
	if uploadPath == "" {
		return nil, fmt.Errorf("something went wrong")
	}

	// Save the uploaded file
	dst, err := os.Create(fmt.Sprintf("%s/%s", fmt.Sprintf("%s/images", os.Getenv("UPLOADS_PATH")), filename))
	if err != nil {
		return nil, fmt.Errorf("something went wrong: %s", err.Error())
	}

	defer dst.Close()

	if _, err := dst.ReadFrom(file); err != nil {
		return nil, fmt.Errorf("something went wrong: %s", err.Error())
	}

	// Open the saved file
	src, err := OpenImage(fmt.Sprintf("%s/%s", fmt.Sprintf("%s/images", os.Getenv("UPLOADS_PATH")), filename), fileType)
	if err != nil {
		return nil, fmt.Errorf("something went wrong: %s", err.Error())
	}

	// Create different quality versions
	err = CreateImageVersions(src, filename, ext, fmt.Sprintf("%s/thumbnails", os.Getenv("UPLOADS_PATH")))
	if err != nil {
		return nil, fmt.Errorf("something went wrong: %s", err.Error())
	}

	return &filename, nil
}
