package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

var (
	tOut = flag.String("t", "odai", "")
)

func main() {
	flag.Parse()
	webhookURL := os.Getenv("PIYOPOKE_WH")

	fmt.Print(*tOut)
	if *tOut == "odai" {
		lines, err := fromFile("pokelist.txt")
		if err != nil {
			log.Fatal(err)
		}

		rand.Seed(time.Now().UnixNano())
		for i := 0; i < 150; i++ {
			rand.Int()
		}

		num := len(lines)
		pokeName := lines[rand.Intn(num)]

		HttpPost(webhookURL, fmt.Sprintf("今日のお題は「%v」です。ボイスチャンネルに入ってください。", pokeName))
	} else if *tOut == "before" {
		HttpPost(webhookURL, "ワンドロスタート！初めてください！ https://mushus.github.io/countdown.html")
	} else if *tOut == "after" {
		HttpPost(webhookURL, "終了ー！ハッシュタグ「#ぴよポケワンドロ」付けてイラストを投稿してください。")
	}
}

func fromFile(filePath string) ([]string, error) {

	f, err := os.Open(filePath)
	if err != nil {
		return []string{}, fmt.Errorf("cannot get file")
	}

	defer f.Close()

	lines := make([]string, 0)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if serr := scanner.Err(); serr != nil {
		return []string{}, fmt.Errorf("cannot read file")
	}

	return lines, nil
}

func HttpPost(url string, text string) error {
	jsonStr := `{"content":"` + text + `"}`

	req, err := http.NewRequest(
		"POST",
		url,
		bytes.NewBuffer([]byte(jsonStr)),
	)
	if err != nil {
		return err
	}

	// Content-Type 設定
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return err
}
