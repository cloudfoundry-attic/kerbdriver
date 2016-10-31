package voltoolshttp_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"fmt"

	"code.cloudfoundry.org/lager/lagertest"
	"github.com/lds-cf/nfsdriver/nfsdriverfakes"
	"github.com/lds-cf/nfsdriver/nfsvoltools"
	"github.com/lds-cf/nfsdriver/nfsvoltools/voltoolshttp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Volman Driver Handlers", func() {

	Context("when generating http handlers", func() {
		var testLogger = lagertest.NewTestLogger("HandlersTest")

		It("should produce a handler with an openPerms route", func() {
			By("faking out the driver")
			voltools := &nfsdriverfakes.FakeVolTools{}
			voltools.OpenPermsReturns(nfsvoltools.ErrorResponse{})
			handler, err := voltoolshttp.NewHandler(testLogger, voltools)
			Expect(err).NotTo(HaveOccurred())

			By("then fake serving the response using the handler")
			route, found := nfsvoltools.Routes.FindRouteByName(nfsvoltools.OpenPermsRoute)
			Expect(found).To(BeTrue())

			path := fmt.Sprintf("http://0.0.0.0%s", route.Path)
			openPermsReq := nfsvoltools.OpenPermsRequest{Opts: map[string]interface{}{"ip": "12.12.12.12"}}
			jsonReq, err := json.Marshal(openPermsReq)
			Expect(err).NotTo(HaveOccurred())
			httpRequest, err := http.NewRequest("POST", path, bytes.NewReader(jsonReq))
			Expect(err).NotTo(HaveOccurred())

			httpResponseRecorder := httptest.NewRecorder()
			handler.ServeHTTP(httpResponseRecorder, httpRequest)

			By("then deserialing the HTTP response")
			response := nfsvoltools.ErrorResponse{}
			body, err := ioutil.ReadAll(httpResponseRecorder.Body)
			err = json.Unmarshal(body, &response)

			By("then expecting correct JSON conversion")
			Expect(err).ToNot(HaveOccurred())
			Expect(response.Err).Should(BeEmpty())
		})

	})
})
