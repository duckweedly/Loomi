package main

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

type versionInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
}

func currentVersion() versionInfo {
	return versionInfo{Version: version, Commit: commit, Date: date}
}
