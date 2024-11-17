package models

import "sync"

type ScanStatus struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Output  string `json:"output"`
}

var ScanStatuses sync.Map
