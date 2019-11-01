package services

import (
	"bytes"
	"compress/gzip"
	"mime/multipart"

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
		api: NewRestAPI(),
	}
}

func (s *StorageService) Upload(fname, uploadCntx string) (string, error) {
	// 1. create destination writer
	bodyBuf := new(bytes.Buffer)
	bodyWriter := multipart.NewWriter(bodyBuf)
	// if fname == "" {
	// 	hashed, err := hashContents(uploadCntx)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	fname = fmt.Sprintf("%x", hashed)
	// }
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

	var storeRes storageResponse
	if err := decode(res.Body, &storeRes); err != nil {
		return "", err
	}

	return storeRes.Hash, nil

}
