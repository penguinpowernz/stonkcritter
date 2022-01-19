package source

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/penguinpowernz/stonkcritter/models"
)

var DisclosuresURLHouse = "https://house-stock-watcher-data.s3-us-west-2.amazonaws.com/data/all_transactions.json"
var DisclosuresURLSenate = "https://senate-stock-watcher-data.s3-us-west-2.amazonaws.com/aggregate/all_transactions.json"

func GetDisclosuresFromFile(fn string) func() ([]models.Disclosure, error) {
	return func() ([]models.Disclosure, error) {
		var v []models.Disclosure
		data, err := ioutil.ReadFile(fn)
		if err != nil {
			return v, err
		}

		err = json.Unmarshal(data, &v)
		return v, err
	}
}

func DownloadDisclosuresFromS3(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		err = errors.New("unexpected status code " + res.Status)
		return nil, err
	}
	log.Println("got data from", url)
	defer res.Body.Close()
	return ioutil.ReadAll(res.Body)
}

func GetDisclosuresFromS3() (dd []models.Disclosure, err error) {
	data, err := DownloadDisclosuresFromS3(DisclosuresURLHouse)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &dd)
	if err != nil {
		return
	}

	data, err = DownloadDisclosuresFromS3(DisclosuresURLSenate)
	if err != nil {
		return
	}
	var _dd []models.Disclosure
	err = json.Unmarshal(data, &_dd)
	dd = append(dd, _dd...)
	return
}
