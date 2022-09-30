package mouselib

import (
	"io/ioutil"

	es8 "github.com/elastic/go-elasticsearch/v8"
)

// es相关

func NewEsClient() (*es8.Client, error) {
	cert, err := ioutil.ReadFile("./http_ca.crt")
	if err != nil {
		return nil, err
	}

	cfg := es8.Config{
		Addresses: []string{"https://localhost:9200"},
		Username:  "elastic",
		Password:  "123456",

		CACert: cert,
	}

	return es8.NewClient(cfg)
}
