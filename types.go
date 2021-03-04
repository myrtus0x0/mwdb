package mwdb

type SampleResp struct {
	Tags         []tag       `json:"tags"`
	LatestConfig interface{} `json:"latest_config"`
	Sha1         string      `json:"sha1"`
	FileSize     int64       `json:"file_size"`
	Sha512       string      `json:"sha512"`
	ID           string      `json:"id"`
	Children     []relative  `json:"children"`
	Sha256       string      `json:"sha256"`
	FileType     string      `json:"file_type"`
	Ssdeep       string      `json:"ssdeep"`
	FileName     string      `json:"file_name"`
	Md5          string      `json:"md5"`
	Parents      []relative  `json:"parents"`
	Type         string      `json:"type"`
	Crc32        string      `json:"crc32"`
	UploadTime   string      `json:"upload_time"`
}

type uploadFileResponse struct {
	Sha1         string        `json:"sha1"`
	FileName     string        `json:"file_name"`
	Tags         []interface{} `json:"tags"`
	Sha512       string        `json:"sha512"`
	FileSize     int64         `json:"file_size"`
	UploadTime   string        `json:"upload_time"`
	Md5          string        `json:"md5"`
	ID           string        `json:"id"`
	Type         string        `json:"type"`
	Ssdeep       string        `json:"ssdeep"`
	Children     []interface{} `json:"children"`
	LatestConfig interface{}   `json:"latest_config"`
	FileType     string        `json:"file_type"`
	Parents      []relative      `json:"parents"`
	Sha256       string        `json:"sha256"`
	Crc32        string        `json:"crc32"`
}

type tag struct {
	Tag string `json:"tag"`
}


type uploadFileMetadata struct {
	Parent   string    `json:"parent"`
	Metakeys []metakey `json:"metakeys"`
	UploadAs string    `json:"upload_as"`
}

type metakey struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}


type configUpload struct {
	Tags       []tag       `json:"tags"`
	ConfigType string      `json:"config_type"`
	Family     string      `json:"family"`
	CFG        interface{} `json:"cfg"`
	ID         string      `json:"id"`
	Parent     string      `json:"parent"`
	Type       string      `json:"type"`
	UploadTime string      `json:"upload_time"`
}

type relative struct {
	Tags       []tag  `json:"tags"`
	Type       string `json:"type"`
	UploadTime string `json:"upload_time"`
	ID         string `json:"id"`
}

