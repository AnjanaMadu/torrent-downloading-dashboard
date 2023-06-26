package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/cenkalti/rain/torrent"
	"github.com/gin-gonic/gin"
)

var (
	ses  *torrent.Session
	zips []*ZipProcess
)

func init() {
	conf := torrent.DefaultConfig
	conf.DataDir = "./downloads"
	ses, _ = torrent.NewSession(conf)

	os.Mkdir("downloads", 0755)
}

func main() {

	// Zip all torrents when they are done downloading
	go func() {
		for {
			for _, tor := range ses.ListTorrents() {
				stats := tor.Stats()
				if stats.Status == torrent.Stopped {
					var found bool
					for _, zp := range zips {
						if zp.ID == tor.ID() {
							found = true
							break
						}
					}
					if !found {
						dPath := path.Join("./downloads", tor.ID())
						outPath := path.Join("./downloads", tor.Name()+".zip")
						zips = append(zips, CreateZip(dPath, outPath, tor.ID()))
					}
				}
			}
			for i, zp := range zips {
				if zp.Status == "done" {
					if len(zips) > 1 {
						zips = append(zips[:i], zips[i+1:]...)
					} else {
						zips = []*ZipProcess{}
					}
					ses.RemoveTorrent(zp.ID)
					os.RemoveAll(zp.InpName)
				}
			}
			time.Sleep(5 * time.Second)
		}
	}()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.StaticFile("/", "index.html")

	r.POST("/api/add", func(c *gin.Context) {
		magnet := c.PostForm("magnet")
		tor, _ := ses.AddURI(magnet, &torrent.AddTorrentOptions{
			StopAfterDownload: true,
		})
		c.JSON(http.StatusOK, tor)
	})

	r.POST("/api/remove", func(c *gin.Context) {
		id := c.PostForm("id")
		err := ses.RemoveTorrent(id)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
		} else {
			c.String(http.StatusOK, "Removed.")
		}
	})

	r.POST("/api/cancelZip", func(c *gin.Context) {
		// id := c.PostForm("id")
		c.String(http.StatusOK, "Removed.")
	})

	r.POST("/api/deleteFile", func(c *gin.Context) {
		name := c.PostForm("name")
		err := os.RemoveAll(path.Join("./downloads", name))
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
		} else {
			c.String(http.StatusOK, "Deleted.")
		}
	})

	r.GET("/api/stats", func(c *gin.Context) {
		var inf GlobalStats
		for _, tor := range ses.ListTorrents() {
			stats := tor.Stats()
			t := Torrent{
				ID:      tor.ID(),
				Name:    tor.Name(),
				Status:  stats.Status.String(),
				Total:   humanBytes(stats.Bytes.Total),
				Current: humanBytes(stats.Bytes.Completed),
				Speed:   humanBytes(int64(stats.Speed.Download)),
			}
			// if stats.ETA != nil {
			// 	t.ETA = stats.ETA.String()
			// }
			inf.Torrents = append(inf.Torrents, t)
		}
		for _, zp := range zips {
			zName := fmt.Sprintf("%s (%d/%d)", path.Base(zp.InpName), zp.Current, zp.Total)
			inf.Files = append(inf.Files, FileFolder{
				Type: "zip",
				Name: zName,
				Size: "",
				ID:   zp.ID,
			})
		}

		files, _ := os.ReadDir("./downloads")
		for _, file := range files {
			if file.IsDir() {
				inf.Files = append(inf.Files, FileFolder{
					Type: "folder",
					Name: path.Base(file.Name()),
					Size: "",
				})
			} else {
				fi, _ := file.Info()
				inf.Files = append(inf.Files, FileFolder{
					Type: "file",
					Name: path.Base(file.Name()),
					Size: humanBytes(fi.Size()),
				})
			}
		}

		c.JSON(http.StatusOK, inf)
	})

	r.GET("/d/:name", func(c *gin.Context) {
		name := c.Param("name")
		c.FileAttachment(path.Join("downloads", name), name)
	})

	r.Run(":8080")
}
