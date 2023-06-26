package main

type Torrent struct {
	ID       string
	Name     string
	Status   string
	Total    string
	Current  string
	Uploaded string
	Speed    string
	ETA      string
}

type FileFolder struct {
	Type string
	Name string
	Size string
	ID   string // for zip
}

type GlobalStats struct {
	Torrents []Torrent
	Files    []FileFolder
}

type ZipProcess struct {
	ID      string
	InpName string
	OutName string
	Status  string
	Current int
	Total   int
}

func CreateZip(inp, out, ID string) *ZipProcess {
	ps := &ZipProcess{ID: ID, InpName: inp, OutName: out}
	go ZipDirectory(inp, out, ps)
	return ps
}
