package support

import (
	p "path"
	"strings"
)

func SplitExt(in string) (path, ext string) {
	dir, file := p.Split(in)
	i := strings.LastIndex(file, ".")
	ext = file[i:]
	fileWithoutExt := file[:i]

	if dir != "" && fileWithoutExt != "" {
		path = p.Join(dir, fileWithoutExt)
	} else if dir != "" {
		path = dir
	} else {
		path = fileWithoutExt
	}

	return
}
