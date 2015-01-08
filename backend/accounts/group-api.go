package accounts

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"

	"github.com/rafael84/go-spa/backend/api"
	"github.com/rafael84/go-spa/backend/context"
	"github.com/rafael84/go-spa/backend/database"
)

func init() {
	api.AddSecureRoute("/accounts/group", GroupHandler)
	api.AddSecureRoute("/accounts/group/{id:[0-9]+}", GroupHandler)
}

func GroupHandler(sc *context.SecureContext, rw http.ResponseWriter, req *http.Request) error {
	var err error

	vars := mux.Vars(req)
	id, found := vars["id"]

	if found {
		var group database.Entity
		group, err = sc.DB.One(&Group{}, "id = $1", id)
		if err != nil {
			log.Errorf("Could not query group id %s: %v", id, err)
			return api.BadRequest(rw, "Could not query group")
		}
		return api.OK(rw, group)
	}

	var groups []database.Entity
	groups, err = sc.DB.Filter(&Group{}, "")
	if err != nil {
		log.Errorf("Could not query groups: %v", err)
		return api.BadRequest(rw, "Could not query groups")
	}
	return api.OK(rw, groups)
}
