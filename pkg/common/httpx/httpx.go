package httpx

import (
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"github.com/fatih/color"
	"github.com/go-resty/resty/v2"
	"github.com/wttech/aemc/pkg/common/pathx"
	"github.com/wttech/aemc/pkg/common/stringsx"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func FileNameFromURL(url string) string {
	return stringsx.BeforeLast(stringsx.AfterLast(url, "/"), "?")
}

type DownloadOpts struct {
	URL      string
	File     string
	Override bool

	AuthToken         string
	AuthBasicUser     string
	AuthBasicPassword string
}

func DownloadWithOpts(opts DownloadOpts) error {
	if len(opts.URL) == 0 {
		return fmt.Errorf("source URL of downloaded file is not specified")
	}
	if len(opts.File) == 0 {
		return fmt.Errorf("destination for downloaded file is not specified")
	}
	if pathx.Exists(opts.File) && !opts.Override {
		return fmt.Errorf("destination for downloaded file already exist")
	}
	client := resty.New()
	client.SetDoNotParseResponse(true)
	if len(opts.AuthBasicUser) > 0 && len(opts.AuthBasicPassword) > 0 {
		client.SetBasicAuth(opts.AuthBasicUser, opts.AuthBasicPassword)
	}
	if len(opts.AuthToken) > 0 {
		client.SetAuthToken(opts.AuthToken)
	}
	fileTmp := opts.File + ".tmp"
	if err := pathx.DeleteIfExists(fileTmp); err != nil {
		return fmt.Errorf("cannot delete temporary file for downloaded from URL '%s' to '%s': %s", opts.URL, opts.File, err)
	}
	defer func() { _ = pathx.DeleteIfExists(fileTmp) }()
	if err := pathx.Ensure(filepath.Dir(fileTmp)); err != nil {
		return err
	}
	res, err := client.R().Get(opts.URL)
	if err != nil {
		return fmt.Errorf("cannot download from URL '%s' to file '%s': %w", opts.URL, opts.File, err)
	}
	defer res.RawBody().Close()
	if res.StatusCode() != http.StatusOK {
		return fmt.Errorf("cannot download from URL '%s' to file '%s': %s", opts.URL, opts.File, res.Status)
	}
	fhTmp, err := os.Create(fileTmp)
	if err != nil {
		return fmt.Errorf("cannot download from URL '%s' as file '%s' cannot be written", opts.URL, opts.File)
	}
	if color.NoColor {
		if _, err := io.Copy(fhTmp, res.RawBody()); err != nil {
			return fmt.Errorf("cannot download from URL '%s' to file '%s': %w", opts.URL, opts.File, err)
		}
	} else {
		bar := pb.Full.Start64(res.RawResponse.ContentLength)
		if _, err := io.Copy(bar.NewProxyWriter(fhTmp), res.RawBody()); err != nil {
			return fmt.Errorf("cannot download from URL '%s' to file '%s': %w", opts.URL, opts.File, err)
		}
		bar.Finish()
	}
	fhTmp.Close()
	err = os.Rename(fileTmp, opts.File)
	if err != nil {
		return fmt.Errorf("cannot move downloaded file from temporary path '%s' to target one '%s': %s", fileTmp, opts.File, err)
	}
	return nil
}

func DownloadWithChanged(opts DownloadOpts) (bool, error) {
	if pathx.Exists(opts.File) && !opts.Override {
		return false, nil
	}
	if err := DownloadWithOpts(opts); err != nil {
		return false, err
	}
	return true, nil
}

func DownloadOnce(url, file string) error {
	_, err := DownloadWithChanged(DownloadOpts{URL: url, File: file})
	return err
}
