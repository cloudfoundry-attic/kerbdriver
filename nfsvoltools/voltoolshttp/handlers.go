package voltoolshttp

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	cf_http_handlers "code.cloudfoundry.org/cfhttp/handlers"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/voldriver/driverhttp"
	"github.com/lds-cf/nfsdriver/nfsvoltools"
	"github.com/tedsuo/rata"
)

func NewHandler(logger lager.Logger, client nfsvoltools.VolTools) (http.Handler, error) {
	logger = logger.Session("server")
	logger.Info("start")
	defer logger.Info("end")

	var handlers = rata.Handlers{
		nfsvoltools.OpenPermsRoute: newOpenPermsHandler(logger, client),
	}

	return rata.NewRouter(nfsvoltools.Routes, handlers)
}

func newOpenPermsHandler(logger lager.Logger, client nfsvoltools.VolTools) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		logger := logger.Session("handle-open-perms")
		logger.Info("start")
		defer logger.Info("end")

		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			logger.Error("failed-reading-request-body", err)
			cf_http_handlers.WriteJSONResponse(w, http.StatusBadRequest, nfsvoltools.ErrorResponse{Err: err.Error()})
			return
		}

		var request nfsvoltools.OpenPermsRequest
		if err = json.Unmarshal(body, &request); err != nil {
			logger.Error("failed-unmarshalling-request-body", err)
			cf_http_handlers.WriteJSONResponse(w, http.StatusBadRequest, nfsvoltools.ErrorResponse{Err: err.Error()})
			return
		}
		ctx := req.Context()
		env := driverhttp.NewHttpDriverEnv(logger, ctx)

		openPermsResponse := client.OpenPerms(env, request)
		if openPermsResponse.Err != "" {
			logger.Error("failed-modifying-permissions", err, lager.Data{"volume": request.Name})
			cf_http_handlers.WriteJSONResponse(w, http.StatusInternalServerError, openPermsResponse)
			return
		}

		cf_http_handlers.WriteJSONResponse(w, http.StatusOK, openPermsResponse)
	}
}
