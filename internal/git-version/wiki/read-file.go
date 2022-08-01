package wiki

import (
	"bufio"
	"github.com/blang/semver/v4"
	"os"
	"strings"
)

func readVersionPage(path string) ([]VersionLine, error) {
	file, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		if os.IsNotExist(err) {
			return []VersionLine{}, nil
		}
		return nil, err
	}

	defer file.Close()

	sc := bufio.NewScanner(file)
	var versions []VersionLine
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "- ") {
			splits := strings.Split(line[2:], " ")
			if len(splits) != 2 {
				continue
			}
			versionStr := splits[0]
			commit := splits[1][1 : len(splits[1])-1]
			v, err := semver.Make(versionStr)
			if err != nil {
				continue
			}
			versions = append(versions, VersionLine{
				Version: v,
				Commit:  commit,
			})
		}
	}

	return versions, nil
}
