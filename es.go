package mouselib

import (
	"io"
	"os"

	es8 "github.com/elastic/go-elasticsearch/v8"
)

// es相关

func NewEsClient() (*es8.Client, error) {
	cert, err := os.Open("./http_ca.crt")
	if err != nil {
		return nil, err
	}
	bs, err := io.ReadAll(cert)
	if err != nil {
		return nil, err
	}

	cfg := es8.Config{
		Addresses: []string{"https://localhost:9200"},
		Username:  "elastic",
		Password:  "123456",

		CACert: bs,
	}

	return es8.NewClient(cfg)
}
