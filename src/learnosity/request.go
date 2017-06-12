package learnosity

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var validSecurityKeys = []string{"consumer_key", "domain", "timestamp", "user_id"}
var validServices = []string{"assess", "author", "data", "items", "questions", "reports", "events"}

//Request Used to generate the necessary security and request data (in the
//correct format) to integrate with any of the Learnosity API services.
type Request struct {
	service         string
	secret          string
	securityPacket  SecurityPacket
	requestPacket   map[string]interface{}
	requestString   string
	action          string
	signRequestData bool
}

//NewRequest generates a new Request to create a signiture
func NewRequest(service string, securityPacket SecurityPacket, secret string, requestPacket *map[string]interface{}) (*Request, error) {
	result := Request{}
	err := result.validateRequiredArgs(service, securityPacket, secret)
	if err != nil {
		return nil, err
	}

	if requestPacket != nil {
		result.requestPacket = *requestPacket
	}

	result.setServiceOptions()
	result.securityPacket.Signature = result.GenerateSignature()

	return &result, nil
}

func (r *Request) setServiceOptions() error {
	if r.service == "assess" || r.service == "questions" {
		r.signRequestData = false
		// The Assess API holds data for the Questions API that includes
		// security information and a signature. Retrieve the security
		// information from $this and generate a signature for the
		// Questions API
		if r.service == "assess" && r.requestPacket != nil && hasKey(r.requestPacket, "questionsApiActivity") {
			questionsAPI := r.requestPacket["questionsApiActivity"].(map[string]interface{})
			domain := "assess.learnosity.com"
			if r.securityPacket.Domain != "" {
				domain = r.securityPacket.Domain
			} else if hasKey(questionsAPI, "domain") {
				domain = questionsAPI["domain"].(string)
			}
			questionsAPI["consumer_key"] = r.securityPacket.ConsumerKey
			questionsAPI["timestamp"] = formatTime(*r.securityPacket.Timestamp)
			questionsAPI["user_id"] = r.securityPacket.UserID

			signatureArray := []string{}
			signatureArray = append(signatureArray, r.securityPacket.ConsumerKey)
			signatureArray = append(signatureArray, domain)
			signatureArray = append(signatureArray, formatTime(*r.securityPacket.Timestamp))
			signatureArray = append(signatureArray, r.securityPacket.UserID)
			signatureArray = append(signatureArray, r.secret)
			questionsAPI["signature"] = r.hashValue(signatureArray)
			r.requestPacket["questionsApiActivity"] = questionsAPI
		}
	} else if r.service == "items" {
		if r.securityPacket.UserID == "" && hasKey(r.requestPacket, "user_id") {
			r.securityPacket.UserID = r.requestPacket["user_id"].(string)
		}
	} else if r.service == "events" {
		r.signRequestData = false
		hashedUsers := map[string]interface{}{}
		if hasKey(r.requestPacket, "users") {
			users := r.requestPacket["users"].([]string)
			for _, u := range users {
				stringToHash := u + r.secret
				h := sha256.New()
				h.Write([]byte(stringToHash))
				userHash := hex.EncodeToString(h.Sum(nil))
				hashedUsers[u] = userHash
			}
			r.requestPacket["users"] = hashedUsers
			bytes, _ := json.Marshal(r.requestPacket)
			r.requestString = string(bytes)
		}
	}

	return nil
}

//Generate creates the data necessary to make a request to Learnosity
func (r *Request) Generate() string {
	output := map[string]interface{}{}
	outputString := ""

	if r.service == "assess" ||
		r.service == "author" ||
		r.service == "data" ||
		r.service == "items" ||
		r.service == "reports" {

		output["security"] = r.securityPacket

		if r.action != "" {
			output["action"] = r.action
		}

		if r.service == "data" {
			bytes, _ := json.Marshal(output)
			return string(bytes)
		}

		if r.requestString != "" {
			output["request"] = r.requestString
		}
		bytes, _ := json.Marshal(output)
		outputString = string(bytes)

	} else if r.service == "questions" {

		//Make map of security packet w/o domain
		output = map[string]interface{}{
			"consumer_key": r.securityPacket.ConsumerKey,
			"timestamp":    formatTime(*r.securityPacket.Timestamp),
			"user_id":      r.securityPacket.UserID,
			"signature":    r.securityPacket.Signature,
		}
		if r.requestString != "" {
			output["request"] = r.requestString
		}
		bytes, _ := json.Marshal(output)
		outputString = string(bytes)

	} else if r.service == "events" {
		output["security"] = r.securityPacket
		if r.requestString != "" {
			output["config"] = r.requestString
		}
		bytes, _ := json.Marshal(output)
		outputString = string(bytes)
	}

	return outputString
}

//GenerateSignature generates a signature hash for the request
func (r *Request) GenerateSignature() string {
	signatureArray := []string{}
	if r.securityPacket.ConsumerKey != "" {
		signatureArray = append(signatureArray, r.securityPacket.ConsumerKey)
	}
	if r.securityPacket.Domain != "" {
		signatureArray = append(signatureArray, r.securityPacket.Domain)
	}
	if r.securityPacket.Timestamp != nil {
		signatureArray = append(signatureArray, formatTime(*r.securityPacket.Timestamp))
	}
	if r.securityPacket.UserID != "" {
		signatureArray = append(signatureArray, r.securityPacket.UserID)
	}

	signatureArray = append(signatureArray, r.secret)

	if r.signRequestData && r.requestString != "" {
		signatureArray = append(signatureArray, r.requestString)
	}

	if r.action != "" {
		signatureArray = append(signatureArray, r.action)
	}

	return r.hashValue(signatureArray)
}

func (r *Request) hashValue(values []string) string {
	valueString := strings.Join(values, "_")
	hash := sha256.New()
	hash.Write([]byte(valueString))
	return hex.EncodeToString(hash.Sum(nil))
}

func (r *Request) validateRequiredArgs(service string, securityPacket SecurityPacket, secret string) error {

	if strings.TrimSpace(service) == "" {
		return errors.New("The `service` argument was empty")
	} else if !containsStr(validServices, service) {
		return fmt.Errorf("Service %s is not valid", service)
	}
	r.service = service

	err := r.validateSecurityPacket(securityPacket)
	if err != nil {
		return err
	}

	if strings.TrimSpace(secret) == "" {
		return errors.New("Must provide valid `secret`")
	}
	r.secret = secret

	return nil
}

func (r *Request) validateSecurityPacket(securityPacket SecurityPacket) error {

	r.securityPacket = securityPacket

	if strings.TrimSpace(securityPacket.ConsumerKey) == "" {
		return errors.New("Consumer key must be provided")
	}

	if r.action == "questions" && strings.TrimSpace(securityPacket.UserID) == "" {
		return errors.New("If using the questions api, a user id needs to be specified")
	}

	if securityPacket.Timestamp == nil {
		now := time.Now()
		securityPacket.Timestamp = &now
	}

	return nil
}
