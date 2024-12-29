package dto

type FileObjectDTO struct {
	ObjectID    string `json:"object_id"`
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	MineType    string `json:"mine_type"`
	DownloadUrl string `json:"download_url"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}
