package torrentStructures

import (
	"github.com/rumanzo/bt2qbt/pkg/fileHelpers"
	"github.com/rumanzo/bt2qbt/pkg/helpers"
	"github.com/rumanzo/bt2qbt/pkg/normalization"
	"sort"
)

func (t *Torrent) IsV2OrHybryd() bool {
	if t.Info.FileTree != nil {
		return true
	}
	return false
}

func (t *Torrent) IsSingle() bool {
	if t.Single != nil {
		return *t.Single
	}
	single := false
	if t.IsV2OrHybryd() {
		// v2 torrents always have at least one file that equal torrent name
		if len(t.Info.FileTree) == 1 {
			torrentName, _ := t.GetNormalizedTorrentName()
			if _, ok := t.Info.FileTree[torrentName]; ok {
				single = true
			}
		}
	} else {
		if t.Info.Files == nil {
			single = true
		}
	}
	t.Single = &single
	return *t.Single
}

// GetFileListWB function that return struct with filelists with bytes from torrent file
func (t *Torrent) GetFileListWB() ([]FilepathLength, bool) {
	if t.FilePathLength == nil {
		if t.IsV2OrHybryd() { // torrents with v2 or hybrid scheme
			result, normalized := getFileListV2(t.Info.FileTree)
			t.FilePathLength = &result
			return *t.FilePathLength, normalized
		} else { // torrent v1 with FileTree
			result, normalized := getFileListV1(t)
			t.FilePathLength = &result
			return *t.FilePathLength, normalized
		}
	} else {
		return *t.FilePathLength, false
	}
}

func (t *Torrent) GetFileList() ([]string, bool) {
	var normalized bool
	if t.FilePathLength == nil {
		_, normalized = t.GetFileListWB()
	}
	if t.FilePaths == nil {
		t.FilePaths = &[]string{}
		for _, fb := range *t.FilePathLength {
			*t.FilePaths = append(*t.FilePaths, fb.Path)
		}
	}
	return *t.FilePaths, normalized
}

func getFileListV1(t *Torrent) ([]FilepathLength, bool) {
	var normalized bool
	var files []FilepathLength
	for _, fileList := range t.Info.Files {

		var normalizedFileList []string
		if fileList.PathUTF8 != nil {
			normalizedFileList = fileList.PathUTF8
		} else {
			normalizedFileList = fileList.Path
		}
		for index, filePathPart := range normalizedFileList {
			normalizedFilePathPart, gotNormalized := normalization.FullNormalize(filePathPart)
			if gotNormalized {
				normalized = true
				normalizedFileList[index] = normalizedFilePathPart
			}
		}
		files = append(files, FilepathLength{
			Path:   fileHelpers.Join(normalizedFileList, `/`),
			Length: fileList.Length,
		})
	}
	return files, normalized
}

func getFileListV2(f interface{}) ([]FilepathLength, bool) {
	var normalized bool
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
			return nfiles, normalized
		}
		s, gotNormalized := getFileListV2(v)
		if gotNormalized {
			normalized = true
		}
		for _, fpl := range s {
			normalizedPath, gotNormalized := normalization.FullNormalize(k)
			if gotNormalized {
				normalized = true
			}
			nfiles = append(nfiles, FilepathLength{Path: fileHelpers.Join(append([]string{normalizedPath}, fpl.Path), `/`), Length: fpl.Length})
		}
	}
	return nfiles, normalized
}

func (t *Torrent) GetTorrentName() string {
	if t.Info.NameUTF8 != "" {
		return t.Info.NameUTF8
	} else {
		return t.Info.Name
	}
}

func (t *Torrent) GetNormalizedTorrentName() (string, bool) {
	torrentName := t.GetTorrentName()
	var normalizedTorrentName string
	var normalized bool
	if fileHelpers.IsAbs(torrentName) {
		normalizedTorrentName, normalized = normalization.NormalizeSpaceEnding(helpers.HandleCesu8(torrentName))
	} else {
		normalizedTorrentName, normalized = normalization.FullNormalize(torrentName)
	}
	return normalizedTorrentName, normalized
}
