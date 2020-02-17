package services

import (
	"bytes"
	"compress/gzip"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"

	"github.com/alabianca/snfs/util"
)

type storageResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Hash    string `json:"data"`
}

type StorageService struct {
	api *RestAPI
}

func NewStroageService() *StorageService {
	return &StorageService{
		api: NewRestAPI(getBaseURL()),
	}
}

func (s *StorageService) Upload(fname, uploadCntx string) (string, error) {
	// 1. create destination writer
	bodyBuf := new(bytes.Buffer)
	bodyWriter := multipart.NewWriter(bodyBuf)
	fileWriter, err := bodyWriter.CreateFormFile("upload", fname)
	if err != nil {
		return "", err
	}
	// 2. create tarball
	gzw := gzip.NewWriter(fileWriter)
	if err := util.WriteTarball(gzw, uploadCntx); err != nil {
		return "", err
	}

	gzw.Close()

	// 3. Upload the file
	contentType := bodyWriter.FormDataContentType()
	url := "v1/storage/fname/" + fname
	bodyWriter.Close()

	// 4. Read response
	res, err := s.api.Post(url, contentType, bodyBuf)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	var storeRes storageResponse
	if err := decode(res.Body, &storeRes); err != nil {
		return "", err
	}

	return storeRes.Hash, nil

}

func (s *StorageService) Download(hash string) error {
	url := "v1/storage/fname/" + hash
	res, err := s.api.Get(url, nil)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.New("Request Failed")
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	hasher := sha1.New()
	io.Copy(hasher, bytes.NewBuffer(bodyBytes))

	sum := fmt.Sprintf("%x", hasher.Sum(nil))
	match := sum == hash
	if !match {
		return errors.New("Hash does not match")
	}

	gzr, _ := gzip.NewReader(bytes.NewBuffer(bodyBytes))
	if err := util.ReadTarball(gzr, hash); err != nil {
		return err
	}

	return nil

}
