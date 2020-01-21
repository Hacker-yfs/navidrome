package engine

import (
	"context"

	"github.com/cloudsonic/sonic-server/consts"
	"github.com/cloudsonic/sonic-server/model"
	"github.com/cloudsonic/sonic-server/utils"
)

type Playlists interface {
	GetAll() (model.Playlists, error)
	Get(id string) (*PlaylistInfo, error)
	Create(ctx context.Context, playlistId, name string, ids []string) error
	Delete(ctx context.Context, playlistId string) error
	Update(ctx context.Context, playlistId string, name *string, idsToAdd []string, idxToRemove []int) error
}

func NewPlaylists(ds model.DataStore) Playlists {
	return &playlists{ds}
}

type playlists struct {
	ds model.DataStore
}

func (p *playlists) Create(ctx context.Context, playlistId, name string, ids []string) error {
	owner := p.getUser(ctx)
	var pls *model.Playlist
	var err error
	// If playlistID is present, override tracks
	if playlistId != "" {
		pls, err = p.ds.Playlist().Get(playlistId)
		if err != nil {
			return err
		}
		if owner != pls.Owner {
			return model.ErrNotAuthorized
		}
		pls.Tracks = nil
	} else {
		pls = &model.Playlist{
			Name:  name,
			Owner: owner,
		}
	}
	for _, id := range ids {
		pls.Tracks = append(pls.Tracks, model.MediaFile{ID: id})
	}

	return p.ds.Playlist().Put(pls)
}

func (p *playlists) getUser(ctx context.Context) string {
	owner := consts.InitialUserName
	user, ok := ctx.Value("user").(*model.User)
	if ok {
		owner = user.UserName
	}
	return owner
}

func (p *playlists) Delete(ctx context.Context, playlistId string) error {
	pls, err := p.ds.Playlist().Get(playlistId)
	if err != nil {
		return err
	}

	owner := p.getUser(ctx)
	if owner != pls.Owner {
		return model.ErrNotAuthorized
	}
	return p.ds.Playlist().Delete(playlistId)
}

func (p *playlists) Update(ctx context.Context, playlistId string, name *string, idsToAdd []string, idxToRemove []int) error {
	pls, err := p.ds.Playlist().Get(playlistId)

	owner := p.getUser(ctx)
	if owner != pls.Owner {
		return model.ErrNotAuthorized
	}

	if err != nil {
		return err
	}
	if name != nil {
		pls.Name = *name
	}
	newTracks := model.MediaFiles{}
	for i, t := range pls.Tracks {
		if utils.IntInSlice(i, idxToRemove) {
			continue
		}
		newTracks = append(newTracks, t)
	}

	for _, id := range idsToAdd {
		newTracks = append(newTracks, model.MediaFile{ID: id})
	}
	pls.Tracks = newTracks

	return p.ds.Playlist().Put(pls)
}

func (p *playlists) GetAll() (model.Playlists, error) {
	return p.ds.Playlist().GetAll(model.QueryOptions{})
}

type PlaylistInfo struct {
	Id        string
	Name      string
	Entries   Entries
	SongCount int
	Duration  int
	Public    bool
	Owner     string
	Comment   string
}

func (p *playlists) Get(id string) (*PlaylistInfo, error) {
	pl, err := p.ds.Playlist().Get(id)
	if err != nil {
		return nil, err
	}

	pinfo := &PlaylistInfo{
		Id:        pl.ID,
		Name:      pl.Name,
		SongCount: len(pl.Tracks), // TODO Use model.Playlist
		Duration:  pl.Duration,
		Public:    pl.Public,
		Owner:     pl.Owner,
		Comment:   pl.Comment,
	}
	pinfo.Entries = make(Entries, len(pl.Tracks))

	// TODO Optimize: Get all tracks at once
	for i, mf := range pl.Tracks {
		mf, err := p.ds.MediaFile().Get(mf.ID)
		if err != nil {
			return nil, err
		}
		pinfo.Entries[i] = FromMediaFile(mf)
	}

	return pinfo, nil
}
