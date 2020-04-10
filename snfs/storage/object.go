package storage

type Object struct {
	Name string `json:"name"`
	Hash string `json:"hash"`
	Size int64  `json:"size"`
	Path string `json:"path"`
}
