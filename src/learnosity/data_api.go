package learnosity

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

//MakeDataRequest calls a Learnosity data api with the provided init data.
//Get initData by calling Init with the proper security and request info
func MakeDataRequest(url string, initData map[string]interface{}) (DataResult, error) {
	result := DataResult{}

	vals := toVals(initData)
	client := http.DefaultClient

	response, err := client.PostForm(url, vals)
	if err != nil {
		return result, err
	}

	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return result, err
	}
	result.String = string(bytes)
	err = json.Unmarshal(bytes, &result.Map)
	return result, err
}

//DataResult ...
type DataResult struct {
	String string
	Map    map[string]interface{}
}

//M alias to map[string]interface{}
type M map[string]interface{}
