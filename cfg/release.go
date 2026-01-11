//go:build release

package cfg

const IsRelease = true
const DBPath = "/srv/apps/releve/shared/data/db.bolt"
const StaticDir = "static/"
const SiteURL = "https://releve.live"
const SiteRoot = "releve.live"
const HLSBaseDir = "/var/www/hls"
const SRSRTMPBase = "rtmp://localhost:1935/live"
