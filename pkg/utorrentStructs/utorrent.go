package utorrentStructs

type ResumeItem struct {
	AddedOn          int64           `bencode:"added_on"`
	Caption          string          `bencode:"caption,omitempty"`
	CompletedOn      int64           `bencode:"completed_on"`
	Downloaded       int64           `bencode:"downloaded"`
	Info             string          `bencode:"info"`
	Label            string          `bencode:"label,omitempty"`
	Labels           []string        `bencode:"labels,omitempty"`
	LastSeenComplete int64           `bencode:"last_seen_complete"`
	Path             string          `bencode:"path"`
	Prio             []byte          `bencode:"prio"`
	Runtime          int64           `bencode:"runtime"`
	Started          int64           `bencode:"started"`
	Targets          [][]interface{} `bencode:"targets,omitempty"`
	Time             int64           `bencode:"time"`
	Trackers         []string        `bencode:"trackers,omitempty"`
	UpSpeed          int64           `bencode:"upspeed"`
	Uploaded         int64           `bencode:"uploaded"`
}
