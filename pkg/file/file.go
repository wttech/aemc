package file

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/wttech/aemc/pkg/common/osx"
	"os"
)

func Download(url, file string) (bool, error) {
	return DownloadWithOpts(DownloadOpts{Url: url, File: file})
}

type DownloadOpts struct {
	Url  string
	File string

	AuthToken         string
	AuthBasicUser     string
	AuthBasicPassword string
}

func DownloadWithOpts(opts DownloadOpts) (bool, error) {
	if len(opts.Url) == 0 {
		return false, fmt.Errorf("source URL of downloaded file is not specified")
	}
	if len(opts.File) == 0 {
		return false, fmt.Errorf("destination for downloaded file is not specified")
	}
	if osx.PathExists(opts.File) {
		return false, nil
	}
	client := resty.New()
	if len(opts.AuthBasicUser) > 0 && len(opts.AuthBasicPassword) > 0 {
		client.SetBasicAuth(opts.AuthBasicUser, opts.AuthBasicPassword)
	}
	if len(opts.AuthToken) > 0 {
		client.SetAuthToken(opts.AuthToken)
	}
	fileTmp := opts.File + ".tmp"
	if osx.PathExists(fileTmp) {
		err := osx.PathDelete(fileTmp)
		if err != nil {
			return false, fmt.Errorf("cannot clean temporary file for downloaded from URL '%s' to '%s': %s", opts.Url, opts.File, err)
		}
	}
	_, err := client.NewRequest().SetOutput(fileTmp).Get(opts.Url)
	if err != nil {
		return false, fmt.Errorf("cannot download file from URL '%s' to '%s': %s", opts.Url, opts.File, err)
	}
	err = os.Rename(fileTmp, opts.File)
	if err != nil {
		return false, fmt.Errorf("cannot move downloaded file from temporary path '%s' to target one '%s': %s", fileTmp, opts.File, err)
	}
	return true, nil
}
