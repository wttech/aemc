package httpx

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/common/stringsx"
	"os"
)

func FileNameFromUrl(url string) string {
	return stringsx.BeforeLast(stringsx.AfterLast(url, "/"), "?")
}

type DownloadOpts struct {
	Url  string
	File string

	AuthToken         string
	AuthBasicUser     string
	AuthBasicPassword string
}

func DownloadWithOpts(opts DownloadOpts) error {
	if len(opts.Url) == 0 {
		return fmt.Errorf("source URL of downloaded file is not specified")
	}
	if len(opts.File) == 0 {
		return fmt.Errorf("destination for downloaded file is not specified")
	}
	if pathx.Exists(opts.File) {
		return fmt.Errorf("destination for downloaded file already exist")
	}
	client := resty.New()
	if len(opts.AuthBasicUser) > 0 && len(opts.AuthBasicPassword) > 0 {
		client.SetBasicAuth(opts.AuthBasicUser, opts.AuthBasicPassword)
	}
	if len(opts.AuthToken) > 0 {
		client.SetAuthToken(opts.AuthToken)
	}
	fileTmp := opts.File + ".tmp"
	if pathx.Exists(fileTmp) {
		err := pathx.Delete(fileTmp)
		if err != nil {
			return fmt.Errorf("cannot clean temporary file for downloaded from URL '%s' to '%s': %s", opts.Url, opts.File, err)
		}
	}
	_, err := client.NewRequest().SetOutput(fileTmp).Get(opts.Url)
	if err != nil {
		return fmt.Errorf("cannot download file from URL '%s' to '%s': %s", opts.Url, opts.File, err)
	}
	err = os.Rename(fileTmp, opts.File)
	if err != nil {
		return fmt.Errorf("cannot move downloaded file from temporary path '%s' to target one '%s': %s", fileTmp, opts.File, err)
	}
	return nil
}

func DownloadWithChanged(opts DownloadOpts) (bool, error) {
	if pathx.Exists(opts.File) {
		return false, nil
	}
	err := DownloadWithOpts(opts)
	if err != nil {
		return false, err
	}
	return true, nil
}

func DownloadOnce(url, file string) error {
	_, err := DownloadWithChanged(DownloadOpts{Url: url, File: file})
	return err
}
