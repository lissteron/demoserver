package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	b, err := json.Marshal([]string{
		"http://google.com",
		"http://google.ru",
		"http://ya.ru",
		"http://yandex.ru",
		"http://mail.ru",
		"https://ya.ru",
		"https://google.com",
		"https://google.ru",
		"https://yandex.ru",
		"https://mail.ru",
	})
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", "http://localhost:8080", bytes.NewBuffer(b))
	if err != nil {
		panic(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("status code: ", resp.StatusCode)

		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	log.Println(string(body))
}
