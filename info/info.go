package info

import (
	"time"
)

var Version = "dev" // default value

type ApiInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Timestamp   string `json:"timestamp"`
}

//GetInfo
func GetInfo() *ApiInfo {
	info := ApiInfo{
		Name:        "LigaPadel API",
		Version:     Version,
		Description: "API para gestionar ligas de pádel por grupos",
		Timestamp:   time.Now().Format(time.RFC3339),
	}
	return &info
}
