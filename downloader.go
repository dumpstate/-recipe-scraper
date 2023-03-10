package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

type ImgDownloader struct {
	Store     *Store
	TargetDir string
}

func download(client *http.Client, url string, fp string) {
	fmt.Printf("Downloading %s to %s\n", url, fp)

	file, err := os.Create(fp)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	defer file.Close()
}

func hash(str string) string {
	hash := md5.Sum([]byte(str))
	return hex.EncodeToString(hash[:])[:20]
}

func NewImgDownloader(store *Store, targetDir string) *ImgDownloader {
	return &ImgDownloader{
		Store:     store,
		TargetDir: targetDir,
	}
}

func (dwnldr *ImgDownloader) DownloadAll() {
	recipes := make(chan *Record[Recipe])
	client := &http.Client{}

	go dwnldr.Store.AllRecipes(recipes)

	for r := range recipes {
		for _, imgUrl := range r.Record.Imgs {
			target := dwnldr.imgPath(r.Id, imgUrl)
			if FileExists(target) {
				continue
			}

			dwnldr.ensureDir(r.Id)
			download(client, imgUrl, target)
		}
	}
}

func (dwnldr *ImgDownloader) ensureDir(recipeId int) {
	dir := dwnldr.recipeDir(recipeId)
	if IsDir(dir) {
		return
	}

	CreateDir(dir)
}

func (dwnldr *ImgDownloader) recipeDir(recipeId int) string {
	return filepath.Join(dwnldr.TargetDir, fmt.Sprintf("%d", recipeId))
}

func (dwnldr *ImgDownloader) imgPath(recipeId int, imgUrl string) string {
	return filepath.Join(
		dwnldr.recipeDir(recipeId),
		fmt.Sprintf("%s%s", hash(imgUrl), path.Ext(imgUrl)),
	)
}
