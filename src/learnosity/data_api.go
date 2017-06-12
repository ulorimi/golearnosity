package learnosity

import (
	"encoding/json"
	"errors"
)

//DataAPI used to make requests to the learnosity data api
type DataAPI struct {
	params         map[string]interface{}
	remote         *Remote
	request        *Request
	url            string
	secJSON        map[string]string
	requestPacket  map[string]interface{}
	requestString  string
	securityPacket SecurityPacket
	action         string
	secret         string
}

//NewDataAPI initializes a new DataAPI with the given url, securityPacket, secret and optional requestPack and action
func NewDataAPI(url string, securityPacket SecurityPacket, secret string, requestPacket *map[string]interface{}, action *string) error {
	api := &DataAPI{}
	api.remote = NewRemote()
	api.url = url
	api.securityPacket = securityPacket
	api.secret = secret
	if action != nil {
		api.action = *action
	}
	if requestPacket != nil {
		api.requestPacket = *requestPacket
	}
	return nil
}

//Request make the post request
func (api *DataAPI) Request() (*Remote, error) {
	api.params = map[string]interface{}{}
	secString := ""
	if api.action == "" {
		request, err := NewRequest("data", api.securityPacket, api.secret, nil)
		if err != nil {
			return nil, err
		}
		api.request = request
		secJSON := map[string]string{}

		secString = api.request.Generate()
		err = json.Unmarshal([]byte(secString), &secJSON)
		if err != nil {
			return nil, err
		}
		api.secJSON = secJSON
	}

	if api.action != "" && api.requestString == "" {
		request, err := NewRequest("data", api.securityPacket, api.secret, nil)
		if err != nil {
			return nil, err
		}
		api.request = request
		secJSON := map[string]string{}
		secString = api.request.Generate()
		err = json.Unmarshal([]byte(secString), &secJSON)
		if err != nil {
			return nil, err
		}
		api.secJSON = secJSON
	}

	if api.action != "" && api.requestString != "" {
		request, err := NewRequest("data", api.securityPacket, api.secret, &api.requestPacket)
		if err != nil {
			return nil, err
		}
		api.request = request
		secJSON := map[string]string{}
		secString = api.request.Generate()
		err = json.Unmarshal([]byte(secString), &secJSON)
		if err != nil {
			return nil, err
		}
		api.params["action"] = api.action
		api.params["request"] = api.requestString
		api.secJSON = secJSON
	}

	api.params["security"] = secString
	api.remote.Post(api.url, api.params)

	return api.remote, nil
}

//RequestJSON returns the result of the data api call as a map
func (api *DataAPI) RequestJSON() (map[string]interface{}, error) {
	_, err := api.Request()
	if err != nil {
		return map[string]interface{}{}, err
	}
	return api.createResponseObject(api.remote)
}

func (api *DataAPI) createResponseObject(remote *Remote) (map[string]interface{}, error) {
	result := map[string]interface{}{}
	if remote == nil {
		return result, errors.New("Remote object was nil")
	}

	result["body"] = remote.Body()
	result["contentType"] = remote.ContentType()
	result["statusCode"] = remote.StatusCode()
	result["timeTaken"] = remote.TimeTaken()

	return result, nil
}
