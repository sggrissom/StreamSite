//go:build !release

package cfg

const IsRelease = false
const DBPath = ".serve/db.bolt"
const StaticDir = ".serve/static/"
const SiteURL = "http://localhost:3000"
const HLSBaseDir = ".serve/hls"
const SRSRTMPBase = "rtmp://localhost:1935/live"
