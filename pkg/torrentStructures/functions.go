package torrentStructures

import (
	"github.com/rumanzo/bt2qbt/pkg/fileHelpers"
)

func (t *Torrent) GetFileList() []string {
	if t.Info.Files != nil {
		return getFileListV1(t)
	} else { // torrent v2 with FileTree
		return getFileListV2(t.Info.FileTree)
	}
}

func getFileListV1(t *Torrent) []string {
	var files []string
	for _, file := range t.Info.Files {
		if file.PathUTF8 != nil {
			files = append(files, fileHelpers.Join(file.PathUTF8, `/`))
		} else {
			files = append(files, fileHelpers.Join(file.Path, `/`))
		}
	}
	return files
}

func getFileListV2(f interface{}) []string {
	nfiles := []string{}

	for k, v := range f.(map[string]interface{}) {
		if len(k) == 0 { // it's means that next will be structure with length and piece root
			return nfiles
		}

		s := getFileListV2(v)

		if len(s) > 0 {
			for _, path := range s {
				nfiles = append(nfiles, fileHelpers.Join(append([]string{k}, path), `/`))
			}
		} else { // it's mean it was last node, just return key
			nfiles = append(nfiles, k)
		}
	}

	return nfiles
}
