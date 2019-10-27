package services

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"fmt"
	"mime/multipart"

	"github.com/alabianca/snfs/util"
)

type storageResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type StorageService struct {
	api *RestAPI
}

func NewStroageService() *StorageService {
	return &StorageService{
		api: NewRestAPI(),
	}
}

func (s *StorageService) Upload(fname, uploadCntx string) error {
	// 1. create destination writer
	bodyBuf := new(bytes.Buffer)
	bodyWriter := multipart.NewWriter(bodyBuf)
	if fname == "" {
		hashed, err := hashContents(uploadCntx)
		fmt.Println("Hashed ", fmt.Sprintf("%x", hashed))
		if err != nil {
			return err
		}

		fname = fmt.Sprintf("%x", hashed)
	}
	fileWriter, err := bodyWriter.CreateFormFile("upload", fname)
	if err != nil {
		return err
	}
	// 2. create tarball
	gzw := gzip.NewWriter(fileWriter)
	if err := util.WriteTarball(gzw, uploadCntx); err != nil {
		return err
	}

	gzw.Close()

	// 3. Upload the file
	contentType := bodyWriter.FormDataContentType()
	url := "v1/storage/fname/" + fname
	bodyWriter.Close()

	// 4. Read response
	res, err := s.api.Post(url, contentType, bodyBuf)
	if err != nil {
		return err
	}

	var storeRes storageResponse
	if err := decode(res.Body, &storeRes); err != nil {
		return err
	}

	fmt.Println(storeRes)

	return nil
}

func hashContents(uploadContents string) ([]byte, error) {
	hash := md5.New()
	gzw := gzip.NewWriter(hash)
	if err := util.WriteTarball(gzw, uploadContents); err != nil {
		return nil, err
	}

	return hash.Sum(nil), nil

}
