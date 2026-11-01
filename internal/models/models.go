package models

type txKeyType string

const (
	RepositoriesTxKey txKeyType = "repositories.tx_key"
)

type LeaseType string

const (
	LeaseTypeAPIServer LeaseType = "api_server"
	LeaseTypeConsumer  LeaseType = "consumer"
)

type NodeInfo struct {
	MachineID string `json:"machine_id"`
	Arch      string `json:"arch"`
	OS        string `json:"os"`
}
