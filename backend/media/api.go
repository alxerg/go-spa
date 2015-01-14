package media

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/gotk/ctx"
	"github.com/gotk/pg"

	"github.com/rafael84/go-spa/backend/base"
)

func init() {
	ctx.Resource("/media", &MediaResource{}, false)
	ctx.Resource("/media/{id:[0-9]+}", &MediaItemResource{}, false)
	ctx.Resource("/media/upload", &MediaUploadResource{}, true)
}

type MediaResource struct {
	*base.Resource
}

func (r *MediaResource) GET(c *ctx.Context, rw http.ResponseWriter, req *http.Request) error {
	medias, err := r.DB(c).FindAll(&Media{}, "")
	if err != nil {
		log.Errorf("Query error: %v", err)
		return ctx.BadRequest(rw, "Query error")
	}
	return ctx.OK(rw, medias)
}

func (r *MediaResource) POST(c *ctx.Context, rw http.ResponseWriter, req *http.Request) error {

	// decode request data
	var form = &struct {
		Name        string `json:"name"`
		MediaTypeId int    `json:"mediaTypeId"`
		LocationId  int    `json:"locationId"`
	}{}
	err := json.NewDecoder(req.Body).Decode(form)
	if err != nil {
		log.Errorf("Could not parse request data: %s", err)
		return ctx.BadRequest(rw, "Could not parse request data")
	}

	// create new media
	path := "/tmp"
	media := &Media{
		Name:        form.Name,
		MediaTypeId: form.MediaTypeId,
		LocationId:  form.LocationId,
		Path:        path,
	}
	err = r.DB(c).Create(media)
	if err != nil {
		log.Errorf("Could not create media %s: %v", form.Name, err)
		return ctx.BadRequest(rw, "Could not create media")
	}

	return ctx.Created(rw, media)
}

type MediaItemResource struct {
	*base.Resource
}

func (r *MediaItemResource) GET(c *ctx.Context, rw http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)
	id := vars["id"]

	media, err := r.DB(c).FindOne(&Media{}, "id = $1", id)
	if err != nil {
		log.Errorf("Could not query media id %s: %v", id, err)
		return ctx.BadRequest(rw, "Could not query media")
	}
	return ctx.OK(rw, media)
}

func (r *MediaItemResource) PUT(c *ctx.Context, rw http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)
	id := vars["id"]

	// decode request data
	var form = &struct {
		Name        string `json:"name"`
		MediaTypeId int    `json:"mediaTypeId"`
		LocationId  int    `json:"locationId"`
		Path        string `json:"path"`
	}{}
	err := json.NewDecoder(req.Body).Decode(form)
	if err != nil {
		log.Errorf("Could not parse request data: %s", err)
		return ctx.BadRequest(rw, "Could not parse request data")
	}

	// get media from database
	var entity pg.Entity
	entity, err = r.DB(c).FindOne(&Media{}, "id = $1", id)
	if err != nil {
		log.Errorf("Could not query media id %s: %v", id, err)
		return ctx.BadRequest(rw, "Could not query media")
	}
	media := entity.(*Media)

	// update the media
	media.Name = form.Name
	media.LocationId = form.LocationId
	media.MediaTypeId = form.MediaTypeId
	media.Path = form.Path
	err = r.DB(c).Update(media)
	if err != nil {
		log.Errorf("Could not edit media %s: %v", form.Name, err)
		return ctx.BadRequest(rw, "Could not edit media")
	}

	return ctx.OK(rw, media)
}

func (r *MediaItemResource) DELETE(c *ctx.Context, rw http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)
	id := vars["id"]

	media, err := r.DB(c).FindOne(&Media{}, "id = $1", id)
	if err != nil {
		log.Errorf("Could not query media id %s: %v", id, err)
		return ctx.BadRequest(rw, "Could not query media")
	}
	err = r.DB(c).Delete(media)
	if err != nil {
		log.Errorf("Could not delete media %s: %v", id, err)
		return ctx.InternalServerError(rw, "Could not delete media")
	}
	return ctx.NoContent(rw)
}

type MediaUploadResource struct {
	*base.Resource
}

func (r *MediaUploadResource) POST(c *ctx.Context, rw http.ResponseWriter, req *http.Request) error {
	reader, err := req.MultipartReader()
	if err != nil {
		return ctx.BadRequest(rw, "Could not upload file")
	}
	var tempFile *os.File
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}

		tempFile, err = ioutil.TempFile("/tmp", "spa")
		if err != nil {
			return ctx.InternalServerError(rw, "Could not create temporary file")
		}
		defer tempFile.Close()

		_, err = io.Copy(tempFile, part)
		if err != nil {
			break
		}
	}
	return ctx.Created(rw, tempFile.Name())
}
