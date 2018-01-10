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
	"github.com/joho/godotenv"
)

var (
	tOut = flag.String("t", "odai", "")
	cfg  config
)

func main() {
	flag.Parse()

	godotenv.Load()

	cfg = config{
		twitterConsumerKey:    os.Getenv("PIYOPOKE_TWITTER_CONSUMER_KEY"),
		twitterConsumerSecret: os.Getenv("PIYOPOKE_TWITTER_CONSUMER_SECRET"),
		twitterAccessToken:    os.Getenv("PIYOPOKE_TWITTER_ACCESS_TOKEN"),
		twitterAccessSecret:   os.Getenv("PIYOPOKE_TWITTER_ACCESS_SECRET"),
		discordWebhook:        os.Getenv("PIYOPOKE_DISCORD_WEBHOOK"),
		pokelistFile:          os.Getenv("PIYOPOKE_POKELIST_FILE"),
	}

	if *tOut == "odai" {
		lines, err := fromFile(cfg.pokelistFile)
		if err != nil {
			log.Fatal(err)
		}

		// 乱数のシードを保存してないので毎回作り直す
		rand.Seed(time.Now().UnixNano())
		for i := 0; i < 150; i++ {
			rand.Int()
		}

		num := len(lines)
		var pokeName string
		for pokeName == "" {
			pokeName = lines[rand.Intn(num)]
		}

		httpPost(cfg.discordWebhook, fmt.Sprintf("今日のお題は「%v」です。ボイスチャンネルに入ってください。", pokeName))
	} else if *tOut == "before" {
		httpPost(cfg.discordWebhook, "ワンドロスタート！始めてください！ https://mushus.github.io/countdown.html")
	} else if *tOut == "after" {
		httpPost(cfg.discordWebhook, "終了ー！ハッシュタグ「#ぴよポケワンドロ」付けてイラストを投稿してください。")
	} else if *tOut == "watch" {
		twitterSearch(cfg.discordWebhook)
	}
}

func twitterSearch(url string) {
	config := oauth1.NewConfig(cfg.twitterConsumerKey, cfg.twitterConsumerSecret)
	token := oauth1.NewToken(cfg.twitterAccessToken, cfg.twitterAccessSecret)

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

type config struct {
	twitterConsumerKey    string
	twitterConsumerSecret string
	twitterAccessToken    string
	twitterAccessSecret   string
	discordWebhook        string
	pokelistFile          string
}
