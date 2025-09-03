package sling

type ResponseData interface {
	IsError() bool
	GetMessage() string
}
