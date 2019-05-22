package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/jhunt/go-ansi"
)

const infoURL = "https://whereslloyd.com/wp-content/themes/whereslloyd/dist/scripts/app-test.js"

type calendarResp struct {
	Items []struct {
		Location string `json:"location"`
		Start    struct {
			DateTime string `json:"dateTime"`
		} `json:"start"`
	} `json:"items"`
}

func main() {
	infoResp, err := http.Get(infoURL)
	if err != nil {
		panic(err.Error())
	}

	infoBody, err := ioutil.ReadAll(infoResp.Body)
	if err != nil {
		panic(err.Error())
	}

	infoResp.Body.Close()
	infoLines := strings.Split(string(infoBody), "\n")

	var apiKey string
	for _, line := range infoLines {
		if strings.Contains(line, "key: ") {
			line = strings.Trim(line, " ")
			line = strings.TrimPrefix(line, "key: ")
			line = strings.Trim(line, "\"")
			apiKey = line
			break
		}
	}

	var calendarHost string
	for _, line := range infoLines {
		if strings.Contains(line, "i = ") {
			line = strings.Trim(line, " ,")
			line = strings.TrimPrefix(line, "i = ")
			line = strings.Trim(line, "\"")
			calendarHost = line
			break
		}
	}

	var calendarPaths []string
	for _, line := range infoLines {
		if strings.Contains(line, "group.calendar.google.com") {
			line = strings.Trim(line, " ,\"")
			calendarPaths = append(calendarPaths, line)
		}
	}

	dateStr := time.Now().Format("2006-01-02")

	data := map[string][]string{}

	for _, path := range calendarPaths {
		url := fmt.Sprintf("%s%s/events?singleEvents=false&timeMin=%[3]sT00:00:00.000Z&timeMax=%[3]sT23:59:59.999Z&key=%[4]s",
			calendarHost,
			path,
			dateStr,
			apiKey,
		)
		//fmt.Println(url)
		calReq, err := http.NewRequest("GET", url, nil)

		if err != nil {
			panic(err.Error())
		}

		calReq.Header.Set("Referer", "https://whereslloyd.com/schedule")

		//fmt.Println(calReq.URL.Path)

		resp, err := http.DefaultClient.Do(calReq)
		if err != nil {
			panic(err.Error())
		}

		respContent, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err.Error())
		}

		resp.Body.Close()

		//fmt.Println(string(respContent))
		respStruct := calendarResp{}

		err = json.Unmarshal(respContent, &respStruct)
		if err != nil {
			panic(err.Error())
		}

		for _, item := range respStruct.Items {
			readableTime, err := time.Parse("2006-01-02T15:04:05-07:00", item.Start.DateTime)
			if err != nil {
				panic(err.Error())
			}

			k := readableTime.In(time.Local).Format(time.Kitchen)
			data[k] = append(data[k], item.Location)
		}
	}

	timeOrder := []string{}

	for time, _ := range data {
		timeOrder = append(timeOrder, time)
	}

	sort.Slice(timeOrder, func(i, j int) bool {
		iT, err := time.Parse(time.Kitchen, timeOrder[i])
		if err != nil {
			panic(err.Error())
		}
		jT, err := time.Parse(time.Kitchen, timeOrder[j])
		if err != nil {
			panic(err.Error())
		}

		return iT.Before(jT)
	})

	for _, t := range timeOrder {
		ansi.Printf("@G{%s}\n", t)
		for _, location := range data[t] {
			ansi.Printf("\t@B{%s}\n", location)
		}
	}
}
