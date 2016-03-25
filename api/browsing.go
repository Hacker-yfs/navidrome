package api

import (
	"fmt"
	"time"

	"github.com/astaxie/beego"
	"github.com/deluan/gosonic/api/responses"
	"github.com/deluan/gosonic/domain"
	"github.com/deluan/gosonic/engine"
	"github.com/deluan/gosonic/utils"
)

type BrowsingController struct {
	BaseAPIController
	browser engine.Browser
}

func (c *BrowsingController) Prepare() {
	utils.ResolveDependencies(&c.browser)
}

func (c *BrowsingController) GetMusicFolders() {
	mediaFolderList, _ := c.browser.MediaFolders()
	folders := make([]responses.MusicFolder, len(mediaFolderList))
	for i, f := range mediaFolderList {
		folders[i].Id = f.Id
		folders[i].Name = f.Name
	}
	response := c.NewEmpty()
	response.MusicFolders = &responses.MusicFolders{Folders: folders}
	c.SendResponse(response)
}

func (c *BrowsingController) GetIndexes() {
	ifModifiedSince := c.ParamTime("ifModifiedSince", time.Time{})

	indexes, lastModified, err := c.browser.Indexes(ifModifiedSince)
	if err != nil {
		beego.Error("Error retrieving Indexes:", err)
		c.SendError(responses.ErrorGeneric, "Internal Error")
	}

	res := responses.Indexes{
		IgnoredArticles: beego.AppConfig.String("ignoredArticles"),
		LastModified:    fmt.Sprint(utils.ToMillis(lastModified)),
	}

	res.Index = make([]responses.Index, len(indexes))
	for i, idx := range indexes {
		res.Index[i].Name = idx.Id
		res.Index[i].Artists = make([]responses.Artist, len(idx.Artists))
		for j, a := range idx.Artists {
			res.Index[i].Artists[j].Id = a.ArtistId
			res.Index[i].Artists[j].Name = a.Artist
		}
	}

	response := c.NewEmpty()
	response.Indexes = &res
	c.SendResponse(response)
}

func (c *BrowsingController) GetMusicDirectory() {
	id := c.RequiredParamString("id", "id parameter required")

	dir, err := c.browser.Directory(id)
	switch {
	case err == domain.ErrNotFound:
		beego.Error("Requested Id", id, "not found:", err)
		c.SendError(responses.ErrorDataNotFound, "Directory not found")
	case err != nil:
		beego.Error(err)
		c.SendError(responses.ErrorGeneric, "Internal Error")
	}

	response := c.NewEmpty()
	response.Directory = c.buildDirectory(dir)
	c.SendResponse(response)
}

func (c *BrowsingController) GetSong() {
	id := c.RequiredParamString("id", "id parameter required")

	song, err := c.browser.GetSong(id)
	switch {
	case err == domain.ErrNotFound:
		beego.Error("Requested Id", id, "not found:", err)
		c.SendError(responses.ErrorDataNotFound, "Directory not found")
	case err != nil:
		beego.Error(err)
		c.SendError(responses.ErrorGeneric, "Internal Error")
	}

	response := c.NewEmpty()
	child := c.ToChild(*song)
	response.Song = &child
	c.SendResponse(response)
}

func (c *BrowsingController) buildDirectory(d *engine.DirectoryInfo) *responses.Directory {
	dir := &responses.Directory{
		Id:         d.Id,
		Name:       d.Name,
		Parent:     d.Parent,
		PlayCount:  d.PlayCount,
		UserRating: d.UserRating,
	}
	if !d.Starred.IsZero() {
		dir.Starred = &d.Starred
	}

	dir.Child = c.ToChildren(d.Entries)
	return dir
}
