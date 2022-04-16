package torrentStructures

type Torrent struct {
	Announce     string           `bencode:"announce"`
	Comment      string           `bencode:"comment"`
	CreatedBy    string           `bencode:"created by"`
	CreationDate int64            `bencode:"creation date"`
	Info         *TorrentInfo     `bencode:"info"`
	Publisher    string           `bencode:"publisher,omitempty"`
	PublisherUrl string           `bencode:"publisher-url,omitempty"`
	PieceLayers  *map[byte][]byte `bencode:"piece layers"`
}

type TorrentInfo struct {
	FileDuration []int64                `bencode:"file-duration,omitempty"`
	FileMedia    []int64                `bencode:"file-media,omitempty"`
	Files        []*TorrentFile         `bencode:"files,omitempty"`
	FileTree     map[string]interface{} `bencode:"file tree,omitempty"`
	Length       int64                  `bencode:"length,omitempty"`
	MetaVersion  int64                  `bencode:"meta version,omitempty"`
	Md5sum       string                 `bencode:"md5sum,omitempty"`
	Name         string                 `bencode:"name,omitempty"`
	NameUTF8     string                 `bencode:"name.utf-8,omitempty"`
	PieceLength  int64                  `bencode:"piece length,omitempty"`
	Pieces       []byte                 `bencode:"pieces,omitempty"`
	Private      uint8                  `bencode:"private,omitempty"`
	Profiles     []*TorrentProfile      `bencode:"profiles,omitempty"`
}

type TorrentFile struct {
	Length   int64    `bencode:"length,omitempty"`
	Md5sum   string   `bencode:"md5sum,omitempty"`
	Path     []string `bencode:"path,omitempty"`
	PathUTF8 []string `bencode:"path.utf-8,omitempty"`
}

type TorrentProfile struct {
	Acodec []byte `bencode:"acodec,omitempty"`
	Height int64  `bencode:"height,omitempty"`
	Vcodec []byte `bencode:"vcodec,omitempty"`
	Width  int64  `bencode:"width,omitempty"`
}
