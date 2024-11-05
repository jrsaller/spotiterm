package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/qeesung/image2ascii/convert"
)

func downloadImage(url string) (string, error) {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Create a temporary file
	tempDir := os.TempDir()
	tempFile, err := os.CreateTemp(tempDir, "image-*.jpg")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	// Write the data to the file
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return "", err
	}

	return tempFile.Name(), nil
}

func GenerateASCIIArt(url string,width int, height int) string {
	// Download the image
	imageFilename, err := downloadImage(url)
	if err != nil {
		fmt.Println("Error downloading image:", err)
		return ""
	}

	// Create convert options
	convertOptions := convert.DefaultOptions
	convertOptions.FixedWidth = width
	convertOptions.FixedHeight = height

	// Create the image converter
	converter := convert.NewImageConverter()
	return converter.ImageFile2ASCIIString(imageFilename, &convertOptions)
}
