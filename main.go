package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/joho/godotenv"
)

var (
	tOut  = flag.String("t", "odai", "")
	debug = flag.Bool("debug", false, "")
	cfg   config
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
		discordWebhookOtona:   os.Getenv("PIYOPOKE_DISCORD_WEBHOOK_OTONA"),
		pokelistFile:          os.Getenv("PIYOPOKE_POKELIST_FILE"),
		pokelogFile:           os.Getenv("PIYOPOKE_POKELOG_FILE"),
		maxPokelog:            os.Getenv("PIYOPOKE_MAX_POKELOG"),
	}

	if *tOut == "odai" {
		maxPokelog, err := strconv.Atoi(cfg.maxPokelog)
		if err != nil {
			log.Fatal(err)
		}

		lines, err := fromFile(cfg.pokelistFile)
		if err != nil {
			log.Fatal(err)
		}

		// 乱数のシードを保存してないので毎回作り直す
		rand.Seed(time.Now().UnixNano())
		for i := 0; i < 150; i++ {
			rand.Int()
		}

		logs, err := fromFile(cfg.pokelogFile)
		if err != nil {
			log.Fatal(err)
		}

		num := len(lines)
		var pokeName string
		for pokeName == "" || indexOf(logs, pokeName) != -1 {
			pokeName = lines[rand.Intn(num)]
		}

		logs = append(logs, pokeName)

		loglen := len(logs)
		firstIdx := loglen - maxPokelog
		if firstIdx < 0 {
			firstIdx = 0
		}

		toFile(cfg.pokelogFile, logs[firstIdx:loglen])

		httpPost(cfg.discordWebhook, fmt.Sprintf("今日のお題は「%v」です。ボイスチャンネルに入ってください。", pokeName))
	} else if *tOut == "before" {
		httpPost(cfg.discordWebhook, "ワンドロスタート！始めてください！")
	} else if *tOut == "after" {
		httpPost(cfg.discordWebhook, "終了ー！ハッシュタグ「#ぴよポケワンドロ」付けてイラストを投稿してください。")
	} else if *tOut == "watch" {
		twitterSearch(cfg.discordWebhook)
	}
}

func twitterSearch(url string) {
	searchWords := []searchWord{
		searchWord{
			word:    "#ぴよポケワンドロ",
			webhook: cfg.discordWebhook,
		},
		searchWord{
			word:    "#おとなのぴよポケワンドロ",
			webhook: cfg.discordWebhookOtona,
		},
	}

	if *debug {
		log.Printf("%v\n", searchWords)
	}
	words := make([]string, len(searchWords))
	for i, v := range searchWords {
		words[i] = v.word
	}

	config := oauth1.NewConfig(cfg.twitterConsumerKey, cfg.twitterConsumerSecret)
	token := oauth1.NewToken(cfg.twitterAccessToken, cfg.twitterAccessSecret)

	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	params := &twitter.StreamFilterParams{
		Track:         words,
		StallWarnings: twitter.Bool(true),
	}

	if *debug {
		log.Printf("start")
	}
	demux := twitter.NewSwitchDemux()
	demux.Tweet = func(tweet *twitter.Tweet) {
		if *debug {
			log.Printf("find tweet")
		}
		if tweet.RetweetedStatus == nil && tweet.QuotedStatus == nil && tweet.ExtendedEntities != nil {
			tweetUrl := fmt.Sprintf("https://twitter.com/%s/status/%s", tweet.User.ScreenName, tweet.IDStr)
			for _, word := range searchWords {
				if strings.Index(tweet.Text, word.word) != -1 {
					if *debug {
						log.Printf("post to discord: %v, %v", word.webhook, tweetUrl)
					}
					httpPost(word.webhook, tweetUrl)
				}
			}
		}
	}

	stream, err := client.Streams.Filter(params)
	if err != nil {
		log.Fatalf("failed to connect filter stream")
	}
	defer stream.Stop()
	go demux.HandleChan(stream.Messages)

	time.Sleep(4 * time.Hour)
	if *debug {
		log.Printf("end watch")
	}
}

func fromFile(filePath string) ([]string, error) {

	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
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

func toFile(filePath string, lines []string) error {
	data := []byte(strings.Join(append(lines, ""), "\n"))
	return ioutil.WriteFile(filePath, data, 0644)
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

func indexOf(haystack []string, needle string) int {
	for i, target := range haystack {
		if target == needle {
			return i
		}
	}
	return -1
}

type config struct {
	twitterConsumerKey    string
	twitterConsumerSecret string
	twitterAccessToken    string
	twitterAccessSecret   string
	discordWebhook        string
	discordWebhookOtona   string
	pokelistFile          string
	pokelogFile           string
	maxPokelog            string
}

type searchWord struct {
	word    string
	webhook string
}
