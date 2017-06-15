package learnosity

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"strings"
	"time"
)

//Init returns initialized object ready for use with Learnosity APIs
func Init(
	service string,
	securityPacket map[string]interface{},
	secret string,
	requestPacket *map[string]interface{},
	action string) (map[string]interface{}, error) {

	output := map[string]interface{}{}

	reqPacket := map[string]interface{}{}

	requestString := ""
	if requestPacket != nil {
		requestString = toJSONString(*requestPacket)
		reqPacket = *requestPacket
	}

	if !hasKey(securityPacket, "timestamp") {
		securityPacket["timestamp"] = formatTime(time.Now())
	}

	if service == "assess" {
		//Insert security info to assess
		reqPacket = insertSecInfoToAssessObject(reqPacket, securityPacket, secret)
	}

	if service == "author" || service == "data" || service == "items" || service == "reports" {
		if hasKey(securityPacket, "user_id") && requestPacket != nil && hasKey(*requestPacket, "user_id") {
			securityPacket["user_id"] = reqPacket["user_id"]
		}
	}

	//Generate signature
	securityPacket["signature"] = generateSignature(service, securityPacket, secret, requestString, action)

	if service == "data" {
		output = map[string]interface{}{
			"security": toJSONString(securityPacket),
			"request":  requestString,
			"action":   action,
		}
	} else if service == "questions" && requestPacket != nil {
		output = extend(securityPacket, *requestPacket)
	} else if service == "assess" {
		return reqPacket, nil
	} else {
		output = map[string]interface{}{
			"security": securityPacket,
			"request":  reqPacket,
		}
	}

	return output, nil
}

func insertSecInfoToAssessObject(requestPacket, securityPacket map[string]interface{}, secret string) map[string]interface{} {

	if hasKey(requestPacket, "questionsApiActivity") {
		questionsAPI := requestPacket["questionsApiActivity"].(map[string]interface{})
		domain := "assess.learnosity.com"
		if hasKey(securityPacket, "domain") {
			domain = securityPacket["domain"].(string)
		} else if hasKey(questionsAPI, "domain") {
			domain = questionsAPI["domain"].(string)
		}

		questionsAPI["consumerKey"] = securityPacket["consumer_key"]
		questionsAPI["timestamp"] = securityPacket["timestamp"]
		questionsAPI["user_id"] = securityPacket["user_id"]
		questionsAPI["signature"] = hashValue([]string{
			securityPacket["consumer_key"].(string),
			domain,
			securityPacket["timestamp"].(string),
			securityPacket["user_id"].(string),
			secret,
		})
		requestPacket["questionsApiActivity"] = questionsAPI
	}

	return requestPacket
}

func generateSignature(service string, packet map[string]interface{}, secret, requestString, action string) string {
	signatureArray := []string{}
	if hasKey(packet, "consumer_key") {
		signatureArray = append(signatureArray, packet["consumer_key"].(string))
	}
	if hasKey(packet, "domain") {
		signatureArray = append(signatureArray, packet["domain"].(string))
	}
	if hasKey(packet, "timestamp") {
		signatureArray = append(signatureArray, packet["timestamp"].(string))
	}
	if hasKey(packet, "user_id") {
		signatureArray = append(signatureArray, packet["user_id"].(string))
	}
	signatureArray = append(signatureArray, secret)

	signData := !(service == "assess" || service == "questions")
	if signData && requestString != "" {
		signatureArray = append(signatureArray, requestString)
	}
	if action != "" {
		signatureArray = append(signatureArray, action)
	}
	return hashValue(signatureArray)
}

func hashValue(values []string) string {
	valueString := strings.Join(values, "_")
	hash := sha256.New()
	hash.Write([]byte(valueString))
	return hex.EncodeToString(hash.Sum(nil))
}

func toJSONString(value map[string]interface{}) string {
	bytes, _ := json.Marshal(value)
	return string(bytes)
}

func extend(base, additional map[string]interface{}) map[string]interface{} {
	for k, val := range additional {
		base[k] = val
	}
	return base
}

func toVals(packet map[string]interface{}) url.Values {
	vals := url.Values{}
	for k := range packet {
		val := packet[k].(string)
		val = strings.Replace(val, "\\", "", -1)
		vals[k] = []string{val}
	}
	return vals
}
