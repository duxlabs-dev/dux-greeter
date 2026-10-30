package greeter

import (
	"crypto/md5"
	"encoding/hex"
	"path/filepath"
	"strings"

	"github.com/duxlabs-dev/dux-greeter/core/internal/utils"
)

// DMS plays video wallpapers itself and caches one still frame per video at
// <cache>/dms/video-wallpaper/<md5 of the path as stored in session.json>.jpg.
// The greeter cannot play video, so it shows that frame instead.
const videoWallpaperFramesName = "video-wallpaper"

var videoWallpaperExtensions = map[string]bool{".mp4": true, ".webm": true, ".mkv": true, ".mov": true, ".m4v": true, ".avi": true}

func VideoWallpaperFramesSource(homeDir string) string {
	return filepath.Join(utils.UserDMSDir(homeDir, ".cache"), videoWallpaperFramesName)
}

func isVideoWallpaper(path string) bool {
	return videoWallpaperExtensions[strings.ToLower(filepath.Ext(path))]
}

func videoWallpaperFrame(framesDir, video string) string {
	sum := md5.Sum([]byte(video))
	return filepath.Join(framesDir, hex.EncodeToString(sum[:])+".jpg")
}
