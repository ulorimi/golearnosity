package learnosity

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

//Remote Used to execute a request to a public endpoint. Useful as a cross domain proxy.
type Remote struct {
	result   map[string]string
	postData map[string]interface{}
	client   *http.Client
	headers  http.Header
}

//NewRemote returns a new Remote instance
func NewRemote() *Remote {
	remote := &Remote{}
	remote.result = map[string]string{}
	remote.postData = map[string]interface{}{}
	return remote
}

//Body returns the body from the completed request
func (r *Remote) Body() string {

	if _, ok := r.result["body"]; ok {
		return r.result["body"]
	}
	return ""
}

//Header returns the header by the given name
func (r *Remote) Header(name string) string {

	if _, ok := r.headers[name]; ok {
		return r.headers[name][0]
	}
	return ""
}

//ContentType returns the content type header of the completed request
func (r *Remote) ContentType() string {
	return r.Header("content_type")
}

//StatusCode returns the body from the completed request
func (r *Remote) StatusCode() string {

	if _, ok := r.result["statusCode"]; ok {
		return r.result["statusCode"]
	}
	return ""
}

//TimeTaken returns
func (r *Remote) TimeTaken() string {
	if _, ok := r.result["total_time"]; ok {
		return r.result["total_time"]
	}
	return ""
}

//Get performs a GET request with optional params and populates the result
func (r *Remote) Get(url string, params *map[string]interface{}) error {
	if params != nil {
		url += "?"
		url += makeQueryString(*params)
	}
	return r.request(url, false)
}

//Post performs a POST request with optional params and populates the result
func (r *Remote) Post(url string, params map[string]interface{}) error {
	r.postData = params
	return r.request(url, true)
}

func (r *Remote) request(url string, post bool) error {
	r.client = http.DefaultClient
	r.client.Timeout = 10000
	start := time.Now().Unix()

	var response *http.Response
	var err error
	if post {
		response, err = r.client.PostForm(url, makeNameValueList(r.postData))
	} else {
		response, err = r.client.Get(url)
	}
	if err != nil {
		return err
	}

	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	took := time.Now().Unix() - start

	r.result["body"] = string(bytes)
	r.result["total_time"] = strconv.Itoa(int(took))
	r.result["statusCode"] = response.Status
	r.headers = response.Header

	return nil
}

func makeQueryString(data map[string]interface{}) string {
	values := makeNameValueList(data)
	return values.Encode()
}

func makeNameValueList(data map[string]interface{}) url.Values {
	vals := url.Values{}
	for k, v := range data {
		vals.Set(k, fmt.Sprintf("%v", v))
	}
	return vals
}
