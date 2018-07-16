package main

import (
	"os"
	"fmt"
	"strings"
	"net/http"

	"github.com/thedevsaddam/gojsonq"
)

type DiscourseClient struct {
	serviceUrl string
}

type Topic struct {
	created string
	url     string
	//author  string
	title   string
}

func (dc DiscourseClient) FetchTopics(dateThreshold string) ([]Topic, error) {
	resp, err := http.Get(dc.serviceUrl + "/latest.json?order=created")
	if err != nil {
		return nil, err
	}

	jsonreader := gojsonq.New().Reader(resp.Body)
	err = jsonreader.Error()
	if err != nil {
                return nil, err
        }
	jsonreader.Macro("s_gt", func(x, y interface{}) (bool, error) {
		xs, okx := x.(string)
		ys, oky := y.(string)
		if !okx || !oky {
			return false, fmt.Errorf("s_gt (string greater than) only supports string")
		}
		return strings.Compare(xs, ys) > 0, nil
	})

	newTopicsQ := jsonreader.From("topic_list.topics").Select("title", "slug", "created_at").Where("created_at", "s_gt", dateThreshold).Get()
	err = jsonreader.Error()
        if err != nil {
                return nil, err
        }

	fmt.Printf("newtopics=%#v\n", newTopicsQ) // debug

	// No new topics
	if newTopicsQ == nil {
		return nil, nil
	}
	newTopics := newTopicsQ.([]interface {})

	res := make([]Topic, len(newTopics))
	for it := 0 ; it < len(newTopics) ; it++ {
		topic := newTopics[it].(map[string]interface {})
		res[it].title = topic["title"].(string)
		res[it].url = dc.serviceUrl + "/t/" + topic["slug"].(string)
		res[it].created = topic["created_at"].(string)
	}

	return res, nil
}

func main() {
	if (len(os.Args)!= 3) {
		os.Exit(1);
	}
	service := os.Args[1]
	threshold := os.Args[2]

	dsClient := DiscourseClient{service}
	topics, err := dsClient.FetchTopics(threshold)
	if (err != nil) {
		fmt.Printf("Error: %#v\n", err)
	}
	fmt.Printf("%#v\n", topics)
}

