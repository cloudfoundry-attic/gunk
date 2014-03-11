package http_client_test

import (
	"fmt"
	. "github.com/cloudfoundry/gunk/http_client"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"time"
)

//We tried to test skipSSLVerification and failed...
//There appears to be a bug in go where a TLS server launched within the same process as the client that makes the connection
//refuses all connections from the client.

func init() {
	net.Listen("tcp", ":8887")

	http.HandleFunc("/sleep", func(w http.ResponseWriter, r *http.Request) {
		sleepTimeInSeconds, _ := strconv.ParseFloat(r.URL.Query().Get("time"), 64)
		time.Sleep(time.Duration(sleepTimeInSeconds * float64(time.Second)))
		fmt.Fprintf(w, "I'm awake!")
	})

	go http.ListenAndServe(":8889", nil)
}

var _ = Describe("HttpClient", func() {
	var client *http.Client

	BeforeEach(func() {
		client = New(true)
	})

	Context("when the request does not time out", func() {
		It("should return the correct response", func() {
			request, _ := http.NewRequest("GET", "http://127.0.0.1:8889/sleep?time=0", nil)
			response, err := client.Do(request)
			Ω(err).ShouldNot(HaveOccurred())
			defer response.Body.Close()
			body, err := ioutil.ReadAll(response.Body)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(string(body)).Should(Equal("I'm awake!"))
		})
	})
})
