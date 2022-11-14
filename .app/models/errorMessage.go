package models

type ErrorMsg struct {
	ErrorCode string `json:"code"`
	Message   string `json:"message"`
}
