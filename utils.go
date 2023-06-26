package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func humanBytes(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}
	if size < 1024*1024 {
		return fmt.Sprintf("%.2f KiB", float64(size)/1024)
	}
	if size < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MiB", float64(size)/(1024*1024))
	}
	return fmt.Sprintf("%.2f GiB", float64(size)/(1024*1024*1024))
}

func ZipDirectory(inputDir, outputZIP string, ps *ZipProcess) error {
	// Calculate total files
	filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ps.Total++
		}
		return nil
	})
	zipFile, err := os.Create(outputZIP)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	filepath.Walk(inputDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			fmt.Println(err)
			return err
		}

		header.Name = filePath
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			fmt.Println(err)
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(filePath)
		if err != nil {
			fmt.Println(err)
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		if err != nil {
			fmt.Println(err)
			return err
		}
		ps.Current++

		return nil
	})

	ps.Status = "done"
	return nil
}
