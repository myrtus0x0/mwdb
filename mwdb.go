package mwdb

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"time"
)

// MwdbClient handles the authentication and server information for MWDB
type MwdbClient struct {
	apiKey   string
	host     string
	protocol string
}

const (
	fileInfoBase   = "/api/file"
	configInfoBase = "/api/config"
)

func (m *MwdbClient) makeAuthenticatedFileUpload(ctx context.Context, contents []byte, URIPath string, parent string, uploadAs string) (string, error) {
	httpCli := http.Client{
		Timeout: 15 * time.Second,
	}

	urlStr := m.protocol + m.host + URIPath

	pseudoFilename := fmt.Sprintf("%x.bin", md5.Sum(contents))

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fw, err := writer.CreateFormFile("file", pseudoFilename)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(fw, bytes.NewBuffer(contents))
	if err != nil {
		return "", err
	}

	if parent != "" {
		optionsw, err := writer.CreateFormField("options")
		if err != nil {
			return "", err
		}
		temp := &uploadFileMetadata{}
		temp.Metakeys = []metakey{}
		temp.Parent = parent
		temp.UploadAs = uploadAs

		marshalledJson, err := json.Marshal(temp)
		if err != nil {
			return "", err
		}

		_, err = io.Copy(optionsw, bytes.NewBuffer(marshalledJson))
		if err != nil {
			return "", err
		}
	}


	err = writer.Close()
	if err != nil {
		return "", err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, body)
	if err != nil {
		return "", err
	}

	request.Header.Add("Content-Type", writer.FormDataContentType())
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.apiKey))

	response, err := httpCli.Do(request)
	if err != nil {
		return "", err
	}

	resp, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	unmarshalledJSON := &uploadFileResponse{}

	err = json.Unmarshal(resp, unmarshalledJSON)
	if err != nil {
		return "", err
	}


	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("invalid response code: %d", response.StatusCode)
	}

	return unmarshalledJSON.Sha256, nil
}

func (m *MwdbClient) makeAuthenticatedHTTPRequest(ctx context.Context, body []byte, method string, URIPath string, headers map[string]string) ([]byte, error) {
	urlStr := m.protocol + m.host + URIPath
	httpCli := http.Client{
		Timeout: 15 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, method, urlStr, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.apiKey))

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := httpCli.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid response code: %d", resp.StatusCode)
	}

	return ioutil.ReadAll(resp.Body)
}

// New returns a instance on a mwdbClient with the appropriate fields set
func New(apikey string, host string, protocol string) (*MwdbClient, error) {
	cli := &MwdbClient{}
	cli.apiKey = apikey
	cli.host = host
	cli.protocol = protocol
	return cli, nil
}

// UploadConfigForSample takes the parent hash and adds a config entry with the given conf argument. Conf should be a struct that gets marshalled into json in this function
func (m *MwdbClient) UploadConfigForSample(ctx context.Context, hash string, conf interface{}, family string) error {
	sampleInfo, err := m.GetInfoAboutSample(ctx, hash)
	if err != nil {
		return err
	}

	confUploadReq := &configUpload{}
	// TODO: what are the implications of this being a empty interface? Everything just gets marshalled?
	confUploadReq.CFG = conf
	confUploadReq.ConfigType = "static"
	confUploadReq.Type = "static_config"
	confUploadReq.Tags = []tag{}
	confUploadReq.Parent = sampleInfo.ID
	confUploadReq.Family = family
	// confUploadReq.UploadTime = sampleInfo.UploadTime

	jsonBody, err := json.MarshalIndent(confUploadReq, "", "    ")
	if err != nil {
		return err
	}

	_, err = m.makeAuthenticatedHTTPRequest(ctx, jsonBody, http.MethodPost, configInfoBase, nil)
	return err
}

func (m *MwdbClient) AddTag(ctx context.Context, sha256 string, tag string) error {
	uriPath := fmt.Sprintf("/api/%s/%s/tag", "file", sha256)
	tagBody := fmt.Sprintf(`{"tag":"%s"}`, tag)
	_, err := m.makeAuthenticatedHTTPRequest(ctx, []byte(tagBody), http.MethodPut, uriPath, nil)
	return err
}

// UploadSample uploads a sample to MWDB with the given tags
func (m *MwdbClient) UploadSample(ctx context.Context, fileContents []byte, tags map[string]string, parent string, uploadAs string) error {
	uriPath := fileInfoBase
	sha256, err := m.makeAuthenticatedFileUpload(ctx, fileContents, uriPath, parent, uploadAs)
	if err != nil {
		return err
	}

	for tagK, tagV := range tags {
		tag := tagK + ":" + tagV
		err := m.AddTag(ctx, sha256, tag)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetInfoAboutSample returns all information about the sample held within MWDB
func (m *MwdbClient) GetInfoAboutSample(ctx context.Context, hash string) (*SampleResp, error) {
	uriPath := fileInfoBase + "/" + hash
	resp, err := m.makeAuthenticatedHTTPRequest(ctx, nil, http.MethodGet, uriPath, nil)
	if err != nil {
		return nil, err
	}

	sampleInfo := &SampleResp{}
	err = json.Unmarshal(resp, sampleInfo)
	if err != nil {
		return nil, err
	}

	return sampleInfo, nil
}

// GetConfigForSample returns the first config instance for the sample if one exists. Otherwise it returns an error of no config found
func (m *MwdbClient) GetConfigForSample(ctx context.Context, hash string) ([]byte, error) {
	sampleInfo, err := m.GetInfoAboutSample(ctx, hash)
	if err != nil {
		return nil, err
	}

	for _, child := range sampleInfo.Children {
		if child.Type == "static_config" {
			uriPath := configInfoBase + "/" + child.ID
			return m.makeAuthenticatedHTTPRequest(ctx, nil, http.MethodGet, uriPath, nil)
		}
	}

	return nil, nil
}
