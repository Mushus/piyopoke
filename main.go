package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	logpkg "log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

var (
	tOut       = flag.String("t", "odai", "")
	configFile = flag.String("c", "config.json", "")
	cfg        Config
	log        *logpkg.Logger
	now        = time.Now
)

func main() {
	flag.Parse()

	data, err := ioutil.ReadFile(*configFile)
	if err != nil {
		logpkg.Fatalf("cannot open config: %v", err)
	}

	// 設定ファイルを読み込む
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		logpkg.Fatalf("invalid config format: %v", err)
	}

	// デバッグ方法
	w := ioutil.Discard
	if cfg.Debug {
		if cfg.LogFile == "" {
			w = os.Stdin
		} else {
			w, err = os.OpenFile(cfg.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				log.Fatalf("falied to open log file: %v", err)
			}
		}
	}
	log = logpkg.New(w, "", logpkg.Ldate|logpkg.Ltime|logpkg.Lshortfile)
	log.Print("start log")

	if *tOut == "odai" {
		// お題
		var pokeName string

		if cfg.CalenderFile != "" {
			f, err := os.Open(cfg.CalenderFile)
			if err == nil {
				pokeName = findToTSV(f)
			}
		}

		// ログファイル読み込み
		logs, err := fromFile(cfg.PokelogFile)
		if err != nil {
			log.Fatal(err)
		}

		// TSVから読み込めなかった場合はランダム
		if pokeName == "" {
			lines, err := fromFile(cfg.PokelistFile)
			if err != nil {
				log.Fatal(err)
			}

			// 乱数のシードを保存してないので毎回作り直す
			rand.Seed(time.Now().UnixNano())
			for i := 0; i < 150; i++ {
				rand.Int()
			}

			num := len(lines)
			for pokeName == "" || indexOf(logs, pokeName) != -1 {
				pokeName = lines[rand.Intn(num)]
				log.Printf("pokename: %v", pokeName)
			}
		}

		// ログに追加
		logs = append(logs, pokeName)
		loglen := len(logs)
		firstIdx := loglen - cfg.MaxPokelog
		if firstIdx < 0 {
			firstIdx = 0
		}

		log.Printf("save pokelog")
		toFile(cfg.PokelogFile, logs[firstIdx:loglen])

		log.Printf("push discord")

		tw := fmt.Sprintf("間もなく開催です！今日のお題は「%v」ですー！#ぴよポケワンドロ", pokeName)
		dc := fmt.Sprintf("今日のお題は「%v」ですー！ボイスチャンネルに入ってください。", pokeName)
		postDefferentText(tw, dc)

	} else if *tOut == "before" {
		post("ワンドロスタート！始めてください！#ぴよポケワンドロ")
	} else if *tOut == "after" {
		post("終了ー！ハッシュタグ「#ぴよポケワンドロ」付けてイラストを投稿してください。#ぴよポケワンドロ")
	} else if *tOut == "watch" {
		twitterSearch(cfg.Discord.Webhook)
	}
}

// 発言する
func post(text string) {
	postDefferentText(text, text)
}

func findToTSV(f io.Reader) string {
	r := csv.NewReader(f)
	r.Comma = '\t'
	records, err := r.ReadAll()
	if err != nil {
		return ""
	}

	for _, v := range records {
		if len(v) < 2 {
			continue
		}
		calenderDate := v[0]
		today := now().Format("01/02")

		if today != calenderDate {
			continue
		}
		chara := v[1]
		return strings.Trim(chara, " ")
	}
	return ""
}

// 別のポケモンに対して発言する
func postDefferentText(twitterText string, discordText string) {
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		httpPost(cfg.Discord.Webhook, discordText)
		wg.Done()
	}()
	go func() {
		tweet(twitterText)
		wg.Done()
	}()
	wg.Wait()
}

func retweet(id int64) {
	tw := cfg.Twitter
	config := oauth1.NewConfig(tw.ConsumerKey, tw.ConsumerSecret)
	token := oauth1.NewToken(tw.AccessToken, tw.AccessSecret)

	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)
	_, _, err := client.Statuses.Retweet(id, nil)
	if err != nil {
		log.Printf("failed to tweet: %v", err)
	}
}
func tweet(text string) {
	tw := cfg.Twitter
	config := oauth1.NewConfig(tw.ConsumerKey, tw.ConsumerSecret)
	token := oauth1.NewToken(tw.AccessToken, tw.AccessSecret)

	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)
	_, _, err := client.Statuses.Update(text, nil)
	if err != nil {
		log.Printf("failed to tweet: %v", err)
	}
}

func twitterSearch(url string) {
	searchWords := []SearchWord{
		SearchWord{
			word:    "#ぴよポケワンドロ",
			webhook: cfg.Discord.Webhook,
			retweet: true,
		},
		SearchWord{
			word:    "#おとなのぴよポケワンドロ",
			webhook: cfg.Discord.WebhookOtona,
			retweet: false,
		},
	}

	log.Printf("%v\n", searchWords)

	words := make([]string, len(searchWords))
	for i, v := range searchWords {
		words[i] = v.word
	}

	tw := cfg.Twitter
	config := oauth1.NewConfig(tw.ConsumerKey, tw.ConsumerSecret)
	token := oauth1.NewToken(tw.AccessToken, tw.AccessSecret)

	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	params := &twitter.StreamFilterParams{
		Track:         words,
		StallWarnings: twitter.Bool(true),
	}

	log.Printf("start")

	demux := twitter.NewSwitchDemux()
	demux.Tweet = func(tweet *twitter.Tweet) {
		log.Printf("find tweet: %v, %v", tweet.User.ScreenName, tweet.IDStr)
		if tweet.RetweetedStatus == nil && tweet.QuotedStatus == nil && tweet.ExtendedEntities != nil {
			tweetURL := fmt.Sprintf("https://twitter.com/%s/status/%s", tweet.User.ScreenName, tweet.IDStr)
			for _, word := range searchWords {
				if strings.Index(tweet.Text, word.word) != -1 {
					log.Printf("post to discord: %v, %v", word.webhook, tweetURL)
					httpPost(word.webhook, tweetURL)
					if word.retweet {
						retweet(tweet.ID)
					}
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

	time.Sleep(2 * time.Hour)
	log.Printf("end watch")
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

type (
	// Twitter の情報
	Twitter struct {
		ConsumerKey    string `json:"consumer_key"`
		ConsumerSecret string `json:"consumer_secret"`
		AccessToken    string `json:"access_token"`
		AccessSecret   string `json:"access_secret"`
	}

	// Discord の情報
	Discord struct {
		Webhook      string `json:"webhook"`
		WebhookOtona string `json:"webhook_otona"`
	}

	// Config 設定
	Config struct {
		Twitter      Twitter `json:"twitter"`
		Discord      Discord `json:"discord"`
		PokelistFile string  `json:"pokelist_file"`
		PokelogFile  string  `json:"pokelog_file"`
		LogFile      string  `json:"log_file"`
		CalenderFile string  `json:"calender_file"`
		MaxPokelog   int     `json:"max_pokelog"`
		Debug        bool    `json:"debug"`
	}

	// SearchWord 検索ワード
	SearchWord struct {
		word    string
		webhook string
		retweet bool
	}
)
