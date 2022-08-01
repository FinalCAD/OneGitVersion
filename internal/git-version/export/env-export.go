package export

import (
	"fmt"
	"github.com/blang/semver/v4"
	"os"
)

func ExportEnv(envPath string, version semver.Version) error {
	file, err := os.OpenFile(envPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0655)
	if err != nil {
		return err
	}

	defer file.Close()

	exportLine := fmt.Sprintf("export FC_VERSION=%s\n", version.String())
	_, err = file.WriteString(exportLine)
	if err != nil {
		return err
	}
	return nil
}
