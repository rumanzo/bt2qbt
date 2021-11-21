package qBittorrentStructures

// https://www.libtorrent.org/manual-ref.html

type QBittorrentFastresume struct {
	ActiveTime                int64          `bencode:"active_time"`
	AddedTime                 int64          `bencode:"added_time"`
	Allocation                string         `bencode:"allocation"`
	ApplyIpFilter             int64          `bencode:"apply_ip_filter"`
	AutoManaged               int64          `bencode:"auto_managed"`
	CompletedTime             int64          `bencode:"completed_time"`
	DisableDht                int64          `bencode:"disable_dht"`
	DisableLsd                int64          `bencode:"disable_lsd"`
	DisablePex                int64          `bencode:"disable_pex"`
	DownloadRateLimit         int64          `bencode:"download_rate_limit"`
	FileFormat                string         `bencode:"file-format"`		// string: "libtorrent resume file"
	FileVersion               int64          `bencode:"file-version"`
	FilePriority              []int64        `bencode:"file_priority"`
	FinishedTime              int64          `bencode:"finished_time"`
	HttpSeeds                 []string       `bencode:"httpseeds"`
	InfoHash                  []byte         `bencode:"info-hash"`		// string, the info hash of the torrent this data is saved for. This is a 20 byte SHA-1 hash of the info section of the torrent if this is a v1 or v1+v2-hybrid torrent.
	InfoHash2                 []byte         `bencode:"info-hash2"`		// string, the v2 info hash of the torrent this data is saved. for, in case it is a v2 or v1+v2-hybrid torrent. This is a 32 byte SHA-256 hash of the info section of the torrent.
	LastDownload              int64          `bencode:"last_download"`
	LastSeenComplete          int64          `bencode:"last_seen_complete"`
	LastUpload                int64          `bencode:"last_upload"`
	LibTorrentVersion         string         `bencode:"libtorrent-version"`
	MappedFiles               []string       `bencode:"mapped_files,omitempty"`
	MaxConnections            int64          `bencode:"max_connections"`
	MaxUploads                int64          `bencode:"max_uploads"`
	NumComplete               int64          `bencode:"num_complete"`
	NumDownloaded             int64          `bencode:"num_downloaded"`
	NumIncomplete             int64          `bencode:"num_incomplete"`
	Paused                    int64          `bencode:"paused"`
	Peers                     int64          `bencode:"peers"`
	Peers6                    int64          `bencode:"peers6"`
	PiecePriority             []byte         `bencode:"piece_priority"`
	Pieces                    []byte         `bencode:"pieces"`		// A string with piece flags, one character per piece. Bit 1 means we have that piece. Bit 2 means we have verified that this piece is correct. This only applies when the torrent is in seed_mode.
	QBtCategory               string         `bencode:"qBt-category"`
	QBtContentLayout          string         `bencode:"qBt-contentLayout"`
	QBtFirstLastPiecePriority string         `bencode:"qBt-firstLastPiecePriority"`
	QbtName                   string         `bencode:"qBt-name"`
	QbtRatioLimit             int64          `bencode:"qBt-ratioLimit"`
	QbtSavePath               string         `bencode:"qBt-savePath"`
	QbtSeedStatus             int64          `bencode:"qBt-seedStatus"`
	QbtSeedingTimeLimit       int64          `bencode:"qBt-seedingTimeLimit"`
	QbtTags                   []string       `bencode:"qBt-tags"`
	SavePath                  string         `bencode:"save_path"`
	SeedMode                  int64          `bencode:"seed_mode"`
	SeedingTime               int64          `bencode:"seeding_time"`
	SequentialDownload        int64          `bencode:"sequential_download"`
	ShareMode                 int64          `bencode:"share_mode"`
	StopWhenReady             int64          `bencode:"stop_when_ready"`
	SuperSeeding              int64          `bencode:"super_seeding"`
	TotalDownloaded           int64          `bencode:"total_downloaded"`	// integer. The number of bytes that have been downloaded in total for this torrent.
	TotalUploaded             int64          `bencode:"total_uploaded"`		// integer. The number of bytes that have been uploaded in total for this torrent.
	Trackers                  [][]string     `bencode:"trackers"`
	UploadMode                int64          `bencode:"upload_mode"`
	UploadRateLimit           int64          `bencode:"upload_rate_limit"`
	UrlList                   int64          `bencode:"url-list"`
	Unfinished                *[]interface{} `bencode:"unfinished,omitempty"`
}
