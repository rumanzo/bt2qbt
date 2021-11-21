package torrentStructures

type Torrent struct {
	Announce     string      `bencode:"announce"`
	Comment      string      `bencode:"comment"`
	CreatedBy    string      `bencode:"created by"`
	CreationDate int64       `bencode:"creation date"`
	Info         TorrentInfo `bencode:"info"`
}

type TorrentInfo struct {
	FileDuration []int64          `bencode:"file-duration,omitempty"`
	FileMedia    []int64          `bencode:"file-media,omitempty"`
	Files        []TorrentFile    `bencode:"files,omitempty"`
	Length       int64            `bencode:"length,omitempty"`
	Md5sum       string           `bencode:"md5sum,omitempty"`
	Name         string           `bencode:"name"`
	PieceLength  int64            `bencode:"piece length"`
	Pieces       []byte           `bencode:"pieces"`
	Private      uint8            `bencode:"private"`
	Profiles     []TorrentProfile `bencode:"profiles,omitempty"`
}

type TorrentFile struct {
	Length int64    `bencode:"Length"`
	Md5sum string   `bencode:"md5sum,omitempty"`
	Path   []string `bencode:"path"`
}

type TorrentProfile struct {
	Acodec []byte `bencode:"acodec"`
	Height int64  `bencode:"height"`
	Vcodec []byte `bencode:"vcodec"`
	Width  int64  `bencode:"width"`
}
