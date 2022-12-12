package repo

import "fmt"

type RepoResult struct {
	Title         string    `json:"title"`
	Path          string    `json:"path"`
	StatusCode    int       `json:"status.code"`
	StatusMessage string    `json:"status.message"`
	Error         RepoError `json:"error"`
}

type RepoError struct {
	Class   string `json:"class"`
	Message string `json:"message"`
}

func (rr RepoResult) IsSuccess() bool {
	return len(rr.Error.Class) == 0 && len(rr.Error.Message) == 0
}

func (rr RepoResult) IsError() bool {
	return !rr.IsSuccess()
}

func (rr RepoResult) ErrorMessage() string {
	return fmt.Sprintf("%s [%d]; %s", rr.Title, rr.StatusCode, rr.StatusMessage)
}
