package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
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
		var pokeName string
		for pokeName == "" {
			pokeName = lines[rand.Intn(num)]
		}

		httpPost(webhookURL, fmt.Sprintf("今日のお題は「%v」です。ボイスチャンネルに入ってください。", pokeName))
	} else if *tOut == "before" {
		httpPost(webhookURL, "ワンドロスタート！始めてください！ https://mushus.github.io/countdown.html")
	} else if *tOut == "after" {
		httpPost(webhookURL, "終了ー！ハッシュタグ「#ぴよポケワンドロ」付けてイラストを投稿してください。")
	} else if *tOut == "watch" {
		twitterSearch(webhookURL)
	}
}

func twitterSearch(url string) {
	consumerKey := os.Getenv("PIYOPOKE_CONSUMER_KEY")
	consumerSecret := os.Getenv("PIYOPOKE_CONSUMER_SECRET")
	accessToken := os.Getenv("PIYOPOKE_ACCESS_TOKEN")
	accessSecret := os.Getenv("PIYOPOKE_ACCESS_SECRET")
	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessToken, accessSecret)

	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	params := &twitter.StreamFilterParams{
		Track:         []string{"#ぴよポケワンドロ"},
		StallWarnings: twitter.Bool(true),
	}

	demux := twitter.NewSwitchDemux()
	demux.Tweet = func(tweet *twitter.Tweet) {
		if tweet.RetweetedStatus == nil && tweet.QuotedStatus == nil && tweet.ExtendedEntities != nil {
			tweetUrl := fmt.Sprintf("https://twitter.com/%s/status/%s", tweet.User.ScreenName, tweet.IDStr)
			httpPost(url, tweetUrl)
		}
	}

	stream, err := client.Streams.Filter(params)
	if err != nil {
		log.Fatalf("failed to connect filter stream")
	}
	go demux.HandleChan(stream.Messages)

	time.Sleep(4 * time.Hour)
	defer stream.Stop()
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

func httpPost(url string, text string) error {
	jsonMap := map[string]string{"content": text}

	b, err := json.Marshal(jsonMap)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		"POST",
		url,
		bytes.NewBuffer(b),
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
