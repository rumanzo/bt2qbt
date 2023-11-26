package torrentStructures

import (
	"github.com/rumanzo/bt2qbt/pkg/fileHelpers"
	"sort"
)

func (t *Torrent) IsV2OrHybryd() bool {
	if t.Info.FileTree != nil {
		return true
	}
	return false
}

func (t *Torrent) IsSingle() bool {
	if t.IsV2OrHybryd() {
		if len(t.Info.FileTree) == 1 {
			var torrentName string
			if t.Info.NameUTF8 != "" {
				torrentName = t.Info.NameUTF8
			} else {
				torrentName = t.Info.Name
			}

			if _, ok := t.Info.FileTree[torrentName]; ok {
				return true
			}
		}
	} else {
		if t.Info.Files == nil {
			return true
		}
	}
	return false
}

// GetFileListWB function that return struct with filelists with bytes from torrent file
func (t *Torrent) GetFileListWB() []FilepathLength {
	if t.FilePathLength == nil {
		if t.IsV2OrHybryd() { // torrents with v2 or hybrid scheme
			result := getFileListV2(t.Info.FileTree)
			t.FilePathLength = &result
			return *t.FilePathLength
		} else { // torrent v1 with FileTree
			result := getFileListV1(t)
			t.FilePathLength = &result
			return *t.FilePathLength
		}
	} else {
		return *t.FilePathLength
	}
}

func (t *Torrent) GetFileList() []string {
	if t.FilePathLength == nil {
		t.GetFileListWB()
	}
	if t.FilePaths == nil {
		t.FilePaths = &[]string{}
		for _, fb := range *t.FilePathLength {
			*t.FilePaths = append(*t.FilePaths, fb.Path)
		}
	}
	return *t.FilePaths
}

func getFileListV1(t *Torrent) []FilepathLength {
	var files []FilepathLength
	for _, file := range t.Info.Files {
		if file.PathUTF8 != nil {
			files = append(files, FilepathLength{
				Path:   fileHelpers.Join(file.PathUTF8, `/`),
				Length: file.Length,
			})
		} else {
			files = append(files, FilepathLength{
				Path:   fileHelpers.Join(file.Path, `/`),
				Length: file.Length,
			})
		}
	}
	return files
}

func getFileListV2(f interface{}) []FilepathLength {
	nfiles := []FilepathLength{}

	// sort map previously
	keys := make([]string, 0, len(f.(map[string]interface{})))
	for k := range f.(map[string]interface{}) {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := f.(map[string]interface{})[k]
		if len(k) == 0 { // it's means that next will be structure with length and piece root
			nfiles = append(nfiles, FilepathLength{Path: "", Length: v.(map[string]interface{})["length"].(int64)})
			return nfiles
		}
		s := getFileListV2(v)
		for _, fpl := range s {
			nfiles = append(nfiles, FilepathLength{Path: fileHelpers.Join(append([]string{k}, fpl.Path), `/`), Length: fpl.Length})
		}
	}
	return nfiles
}
