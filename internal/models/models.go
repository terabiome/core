package models

type LeaseType string

const (
	LeaseTypeAPIServer LeaseType = "api_server"
	LeaseTypeConsumer  LeaseType = "consumer"
)
