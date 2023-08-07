package models

import "encoding/json"

type ExarotonBody struct {
	Success bool            `json:"success"`
	Error   string          `json:"error"`
	Data    json.RawMessage `json:"data"`
}

type ExarotonServer struct {
	Id       string                 `json:"id"`
	Name     string                 `json:"name"`
	Address  string                 `json:"address"`
	Motd     string                 `json:"motd"`
	Status   int                    `json:"status"`
	Host     string                 `json:"host"`
	Port     int                    `json:"port"`
	Software ExarotonServerSoftware `json:"software"`
}

type ExarotonServerSoftware struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
}
