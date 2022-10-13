package mouselib_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	es8 "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/mouseleee/mouselib"
)

type TestDesc struct {
	Id   int    `json:"id"`
	Text string `json:"text"`
}

func TestNewEsClientCreate(t *testing.T) {
	t.FailNow()
	es, err := mouselib.NewEsClient()
	if err != nil {
		t.Error(err)
		return
	}

	var r map[string]interface{}

	rsp, err := es.Info()
	if err != nil {
		t.Error(err)
		return
	}
	defer rsp.Body.Close()

	if rsp.IsError() {
		t.Error(err)
		return
	}

	if err = json.NewDecoder(rsp.Body).Decode(&r); err != nil {
		t.Error(err)
	}

	t.Logf("Client: %s", es8.Version)
	t.Logf("Server: %s", r["version"].(map[string]interface{})["number"])
	t.Log(strings.Repeat("~", 37))

	testInst := TestDesc{
		Id:   6,
		Text: "mewo",
	}
	rd := esutil.NewJSONReader(testInst)
	req := esapi.IndexRequest{
		Index:      "test",
		DocumentID: "test",
		Body:       rd,
		Refresh:    "true",
	}

	rsp, err = req.Do(context.Background(), es)
	if err != nil {
		t.Error(err)
		return
	}
	defer rsp.Body.Close()

	if rsp.IsError() {
		t.Logf("[%s] Error indexing document ID=%s", rsp.Status(), "test")
	} else {
		// Deserialize the response into a map.
		if err := json.NewDecoder(rsp.Body).Decode(&r); err != nil {
			t.Logf("Error parsing the response body: %s", err)
		} else {
			// Print the response status and indexed document version.
			t.Logf("[%s] %s; version=%d", rsp.Status(), r["result"], int(r["_version"].(float64)))
		}
	}
}

func TestNewEsClientGet(t *testing.T) {
	es, err := mouselib.NewEsClient()
	if err != nil {
		t.Error(err)
		return
	}

	var r map[string]interface{}

	// 	PUT /<target>/_doc/<_id>
	// POST /<target>/_doc/
	// PUT /<target>/_create/<_id>
	// POST /<target>/_create/<_id>

	req := esapi.SearchRequest{
		Query: "text",
	}

	rsp, err := req.Do(context.Background(), es)
	if err != nil {
		t.Error(err)
		return
	}
	defer rsp.Body.Close()

	if rsp.IsError() {
		t.Logf("[%s] Error search index ID=%s", rsp.Status(), "test")
	} else {
		// Deserialize the response into a map.
		if err := json.NewDecoder(rsp.Body).Decode(&r); err != nil {
			t.Logf("Error parsing the response body: %s", err)
		} else {
			// Print the response status and indexed document version.
			t.Logf("[status %s] all:%v", rsp.Status(), r["hits"])
		}
	}
}
