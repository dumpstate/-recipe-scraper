package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
)

type ImgProcessor struct {
	SourceDir   string
	TargetDir   string
	Concurrency int
}

type SubImager interface {
	SubImage(r image.Rectangle) image.Image
}

func NewImgProcessor(sourceDir string, targetDir string, concurrency int) *ImgProcessor {
	if !IsDir(sourceDir) {
		log.Fatal(fmt.Sprintf("source directory %s does not exist", sourceDir))
	}

	return &ImgProcessor{
		SourceDir:   sourceDir,
		TargetDir:   targetDir,
		Concurrency: concurrency,
	}
}

func tree(dir string, fs chan<- string, root bool) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		path := filepath.Join(dir, f.Name())

		if f.IsDir() {
			tree(path, fs, false)
		} else {
			fs <- path
		}
	}

	if root {
		close(fs)
	}
}

func cropSquare(targetDir string, imgPath string) {
	fmt.Printf("Processing %s\n", imgPath)

	fname := filepath.Base(imgPath)
	imgf, err := os.Open(imgPath)
	if err != nil {
		log.Fatal(err)
	}
	defer imgf.Close()

	ext := filepath.Ext(imgPath)
	var dimg image.Image
	if ext == ".png" {
		dimg, err = png.Decode(imgf)
		if err != nil {
			log.Fatal(err)
		}
	} else if ext == ".jpg" {
		dimg, err = jpeg.Decode(imgf)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatalf("Unsupported format: %s", imgPath)
	}

	bounds := dimg.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	size := int(math.Min(float64(w), float64(h)))
	xoff, yoff := (w-size)/2, (h-size)/2
	crop := image.Rect(xoff, yoff, size+xoff, size+yoff)
	cropped := dimg.(SubImager).SubImage(crop)

	tdir := filepath.Join(targetDir, filepath.Base(filepath.Dir(imgPath)))
	CreateDir(tdir)
	basename := fmt.Sprintf("%s.png", fname[:len(fname)-len(ext)])
	out, err := os.Create(filepath.Join(tdir, basename))
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	png.Encode(out, cropped)
}

func cropWorker(targetDir string, fs <-chan string) {
	for f := range fs {
		cropSquare(targetDir, f)
	}
}

func (imgp *ImgProcessor) CropSquareAll() {
	fs := make(chan string, imgp.Concurrency)

	go tree(imgp.SourceDir, fs, true)

	for w := 1; w <= imgp.Concurrency-1; w++ {
		go cropWorker(imgp.TargetDir, fs)
	}

	cropWorker(imgp.TargetDir, fs)
}
