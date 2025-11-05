//go:build release

package cfg

const IsRelease = true
const DBPath = "/var/lib/stream/data/db.bolt"
const StaticDir = "static/"
const SiteURL = "https://releve.live"
const SiteRoot = "releve.live"
const HLSBaseDir = "/var/www/hls"
const SRSRTMPBase = "rtmp://localhost:1935/live"
