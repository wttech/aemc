package certx

import (
	"encoding/pem"
	"fmt"
	"os"
)

func CreateTmpDerFileBasedOnPem(block *pem.Block) (*string, func(), error) {
	tempDerFile, err := os.CreateTemp("", "tmp-certx-*.der")

	if err != nil {
		return nil, nil, fmt.Errorf("failed to create temp file for storing DER certificate: %w", err)
	}

	err = writeCertToDer(tempDerFile, block)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to write DER certificate: %w", err)
	}

	tmpFileName := tempDerFile.Name()

	return &tmpFileName, func() { os.Remove(tmpFileName) }, nil
}

func writeCertToDer(tempDerFile *os.File, pemBlock *pem.Block) error {
	if _, err := tempDerFile.Write(pemBlock.Bytes); err != nil {
		return err
	}
	err := tempDerFile.Close()
	if err != nil {
		return err
	}
	return nil
}
