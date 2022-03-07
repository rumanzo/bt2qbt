package qBittorrentStructures

import "github.com/rumanzo/bt2qbt/pkg/torrentStructures"

// https://www.libtorrent.org/manual-ref.html

type QBittorrentFastresume struct {
	ActiveTime                int64                            `bencode:"active_time"` // integer. The number of seconds this torrent has been active. i.e. not paused.
	AddedTime                 int64                            `bencode:"added_time"`
	Allocation                string                           `bencode:"allocation"`      // The allocation mode for the storage. Can be either allocate or sparse.
	ApplyIpFilter             int64                            `bencode:"apply_ip_filter"` // integer. 1 if the torrent_flags::apply_ip_filter is set.
	AutoManaged               int64                            `bencode:"auto_managed"`    // integer. 1 if the torrent is auto managed, otherwise 0.
	BannedPeers               []byte                           `bencode:"banned_peers"`    // string. This string has the same format as peers but instead represent IPv4 peers that we have banned.
	BannedPeers6              []byte                           `bencode:"banned_peers6"`   // string. This string has the same format as peers6 but instead represent IPv6 peers that we have banned.
	CompletedTime             int64                            `bencode:"completed_time"`
	DisableDht                int64                            `bencode:"disable_dht"`         // integer. 1 if the torrent_flags::disable_dht is set.
	DisableLsd                int64                            `bencode:"disable_lsd"`         // integer. 1 if the torrent_flags::disable_lsd is set.
	DisablePex                int64                            `bencode:"disable_pex"`         // integer. 1 if the torrent_flags::disable_pex is set.
	DownloadRateLimit         int64                            `bencode:"download_rate_limit"` // integer. The download rate limit for this torrent in case one is set, in bytes per second.
	FileFormat                string                           `bencode:"file-format"`         // string: "libtorrent resume file"
	FilePriority              []int64                          `bencode:"file_priority"`       // list of integers. One entry per file in the torrent. Each entry is the priority of the file with the same index.
	FileVersion               int64                            `bencode:"file-version"`
	FinishedTime              int64                            `bencode:"finished_time"`
	HttpSeeds                 []string                         `bencode:"httpseeds"`      // list of strings. List of HTTP seed URLs used by this torrent. The URLs are expected to be properly encoded and not contain any illegal url characters.
	Info                      []*torrentStructures.TorrentInfo `bencode:"info,omitempty"` // If this field is present, it should be the info-dictionary of the torrent this resume data is for. Its SHA-1 hash must match the one in the info-hash field. When present, the torrent is loaded from here, meaning the torrent can be added purely from resume data (no need to load the .torrent file separately). This may have performance advantages.
	InfoHash                  []byte                           `bencode:"info-hash"`      // string, the info hash of the torrent this data is saved for. This is a 20 byte SHA-1 hash of the info section of the torrent if this is a v1 or v1+v2-hybrid torrent.
	InfoHash2                 []byte                           `bencode:"info-hash2"`     // string, the v2 info hash of the torrent this data is saved. for, in case it is a v2 or v1+v2-hybrid torrent. This is a 32 byte SHA-256 hash of the info section of the torrent.
	LastDownload              int64                            `bencode:"last_download"`  // integer. The number of seconds since epoch when we last downloaded payload from a peer on this torrent.
	LastSeenComplete          int64                            `bencode:"last_seen_complete"`
	LastUpload                int64                            `bencode:"last_upload"` // integer. The number of seconds since epoch when we last uploaded payload to a peer on this torrent.
	LibTorrentVersion         string                           `bencode:"libtorrent-version"`
	MappedFiles               []string                         `bencode:"mapped_files,omitempty"` // list of strings. If any file in the torrent has been renamed, this entry contains a list of all the filenames. In the same order as in the torrent file.
	MaxConnections            int64                            `bencode:"max_connections"`        // integer. The max number of peer connections this torrent may have, if a limit is set.
	MaxUploads                int64                            `bencode:"max_uploads"`            // integer. The max number of unchoked peers this torrent may have, if a limit is set.
	NumComplete               int64                            `bencode:"num_complete"`
	NumDownloaded             int64                            `bencode:"num_downloaded"`
	NumIncomplete             int64                            `bencode:"num_incomplete"`
	Paused                    int64                            `bencode:"paused"`         // 	integer. 1 if the torrent is paused, 0 otherwise.
	Peers                     int64                            `bencode:"peers"`          // string. This string contains IPv4 and port pairs of peers we were connected to last session. The endpoints are in compact representation. 4 bytes IPv4 address followed by 2 bytes port. Hence, the length of this string should be divisible by 6.
	Peers6                    int64                            `bencode:"peers6"`         // 	string. This string contains IPv6 and port pairs of peers we were connected to last session. The endpoints are in compact representation. 16 bytes IPv6 address followed by 2 bytes port. The length of this string should be divisible by 18.
	PiecePriority             []byte                           `bencode:"piece_priority"` // string of bytes. Each byte is interpreted as an integer and is the priority of that piece.
	Pieces                    []byte                           `bencode:"pieces"`         // A string with piece flags, one character per piece. Bit 1 means we have that piece. Bit 2 means we have verified that this piece is correct. This only applies when the torrent is in seed_mode.
	QBtCategory               string                           `bencode:"qBt-category"`
	QBtContentLayout          string                           `bencode:"qBt-contentLayout"` // Original, Subfolder, NoSubfolder
	QBtFirstLastPiecePriority string                           `bencode:"qBt-firstLastPiecePriority"`
	QbtName                   string                           `bencode:"qBt-name"`
	QbtRatioLimit             int64                            `bencode:"qBt-ratioLimit"`
	QbtSavePath               string                           `bencode:"qBt-savePath"`
	QbtSeedStatus             int64                            `bencode:"qBt-seedStatus"`
	QbtSeedingTimeLimit       int64                            `bencode:"qBt-seedingTimeLimit"`
	QbtTags                   []string                         `bencode:"qBt-tags"`
	SavePath                  string                           `bencode:"save_path"`           // string. The save path where this torrent was saved. This is especially useful when moving torrents with move_storage() since this will be updated.
	SeedMode                  int64                            `bencode:"seed_mode"`           // integer. 1 if the torrent is in seed mode, 0 otherwise.
	SeedingTime               int64                            `bencode:"seeding_time"`        // integer. The number of seconds this torrent has been active and seeding.
	SequentialDownload        int64                            `bencode:"sequential_download"` // integer. 1 if the torrent is in sequential download mode, 0 otherwise.
	ShareMode                 int64                            `bencode:"share_mode"`          // integer. 1 if the torrent_flags::share_mode is set.
	StopWhenReady             int64                            `bencode:"stop_when_ready"`     // integer. 1 if the torrent_flags::stop_when_ready is set.
	SuperSeeding              int64                            `bencode:"super_seeding"`       // integer. 1 if the torrent_flags::super_seeding is set.
	TotalDownloaded           int64                            `bencode:"total_downloaded"`    // integer. The number of bytes that have been downloaded in total for this torrent.
	TotalUploaded             int64                            `bencode:"total_uploaded"`      // integer. The number of bytes that have been uploaded in total for this torrent.
	Trackers                  [][]string                       `bencode:"trackers"`            // list of lists of strings. The top level list lists all tracker tiers. Each second level list is one tier of trackers.
	Unfinished                *[]interface{}                   `bencode:"unfinished,omitempty"`
	UploadMode                int64                            `bencode:"upload_mode"`       // integer. 1 if the torrent_flags::upload_mode is set.
	UploadRateLimit           int64                            `bencode:"upload_rate_limit"` // integer. In case this torrent has a per-torrent upload rate limit, this is that limit. In bytes per second.
	UrlList                   int64                            `bencode:"url-list"`          // list of strings. List of url-seed URLs used by this torrent. The URLs are expected to be properly encoded and not contain any illegal url characters.
}
