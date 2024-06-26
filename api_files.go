package moonshot

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/northes/go-moonshot/enum"
	"github.com/northes/gox/httpx"
)

type files struct {
	client *httpx.Client
}

func (c *Client) Files() *files {
	return &files{
		client: c.newHTTPClient(),
	}
}

type FilesUploadRequest struct {
}
type FilesUploadResponse struct {
}

func (f *files) Upload(filePath string) (*FilesUploadResponse, error) {
	const path = "/v1/files"

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileField, err := writer.CreateFormFile("file", filepath.Base(file.Name()))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(fileField, file)
	if err != nil {
		return nil, err
	}

	err = writer.WriteField("purpose ", enum.FilePurposeExtract.String())
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	resp, err := f.client.AddPath(path).SetBody(body).SetContentType(writer.FormDataContentType()).Post()
	if err != nil {
		return nil, err
	}

	uploadResponse := new(FilesUploadResponse)
	err = resp.Unmarshal(uploadResponse)
	if err != nil {
		return nil, err
	}

	return uploadResponse, nil
}

type FilesListRequest struct {
}
type FilesListResponse struct {
	Object string                  `json:"object"`
	Data   []FilesListResponseData `json:"data"`
}
type FilesListResponseData struct {
	ID           string            `json:"id"`
	Object       string            `json:"object"`
	Bytes        int64             `json:"bytes"`
	CreatedAt    int64             `json:"created_at"`
	Filename     string            `json:"filename"`
	Purpose      enum.FilesPurpose `json:"purpose"`
	Status       string            `json:"status"`
	StatusDetail string            `json:"status_detail"`
}

func (f *files) Lists() (*FilesListResponse, error) {
	const path = "/v1/files"
	resp, err := f.client.AddPath(path).Get()
	if err != nil {
		return nil, err
	}
	listResponse := new(FilesListResponse)
	err = resp.Unmarshal(listResponse)
	if err != nil {
		return nil, err
	}
	return listResponse, nil
}

type FilesDeleteResponse struct {
	Deleted bool   `json:"deleted"`
	Id      string `json:"id"`
	Object  string `json:"object"`
}

func (f *files) Delete(fileID string) (*FilesDeleteResponse, error) {
	const path = "/v1/files/%s"
	fullPath := fmt.Sprintf(path, fileID)
	resp, err := f.client.AddPath(fullPath).Delete()
	if err != nil {
		return nil, err
	}
	deleteResponse := new(FilesDeleteResponse)
	err = resp.Unmarshal(deleteResponse)
	if err != nil {
		return nil, err
	}
	return deleteResponse, nil
}

type FilesInfoResponse struct {
	Id            string `json:"id"`
	Object        string `json:"object"`
	Bytes         int    `json:"bytes"`
	CreatedAt     int    `json:"created_at"`
	Filename      string `json:"filename"`
	Purpose       string `json:"purpose"`
	Status        string `json:"status"`
	StatusDetails string `json:"status_details"`
}

func (f *files) Info(fileID string) (*FilesInfoResponse, error) {
	const path = "/v1/files/%s"
	fullPath := fmt.Sprintf(path, fileID)
	resp, err := f.client.AddPath(fullPath).Get()
	if err != nil {
		return nil, err
	}
	infoResponse := new(FilesInfoResponse)
	err = resp.Unmarshal(infoResponse)
	if err != nil {
		return nil, err
	}
	return infoResponse, nil
}

type FileContentResponse struct {
	Content  string `json:"content"`
	FileType string `json:"file_type"`
	Filename string `json:"filename"`
	Title    string `json:"title"`
	Type     string `json:"type"`
}

func (f *files) Content(fileID string) (*FileContentResponse, error) {
	const path = "/v1/files/%s/content"
	fullPath := fmt.Sprintf(path, fileID)
	resp, err := f.client.AddPath(fullPath).Get()
	if err != nil {
		return nil, err
	}
	contentResponse := new(FileContentResponse)
	err = resp.Unmarshal(contentResponse)
	if err != nil {
		return nil, err
	}
	return contentResponse, nil
}
