package sources

// TODO call this summoner.go

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	log "github.com/Sirupsen/logrus"
)

type RequestError struct {
	Message      string
	ResponseCode int
}

func (r RequestError) Error() string {
	return fmt.Sprintf("%s, response code: %d", r.Message, r.ResponseCode)
}

// GetJSON abstracts away from the usual log/connect/retry logic involving GET
// requests. This particular version unmarshals JSON.
func GetJSON(req *http.Request, target interface{}) error {
	return get(req, target, json.Unmarshal)
}

// GetJSON abstracts away from the usual log/connect/retry logic involving GET
// requests. This particular version unmarshals XML.
func GetXML(req *http.Request, target interface{}) error {
	return get(req, target, xml.Unmarshal)
}

// get can be used to generically call URLs and deserialize the results.
func get(req *http.Request, target interface{}, unmarshalFunc func([]byte, interface{}) error) error {
	log.WithFields(
		log.Fields{
			"URL": req.URL.String(),
		}).Debug("Request")

	// Be nice and tell them who we are.
	req.Header.Set("User-Agent", "github.com/haarts/getme")

	resp, err := http.DefaultClient.Do(req)

	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()
	if err != nil {
		log.WithFields(
			log.Fields{
				"error": err,
				"URL":   req.URL.String(),
			}).Error("GET error")
		return err //TODO retry a couple of times when it's a timeout.
	}

	if resp.StatusCode != 200 {
		log.WithFields(
			log.Fields{
				"code": strconv.Itoa(resp.StatusCode),
				"URL":  req.URL.String(),
			}).Warn("Non 200 response code")
		return RequestError{
			Message:      "Search returned non 200 status code",
			ResponseCode: resp.StatusCode,
		}

	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = unmarshalFunc(body, target)
	if err != nil {
		return err
	}

	return nil
}
