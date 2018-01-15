# ぴよポケbot

ぴよぽけさん「今日のお題は「イーブイ」です。ボイスチャンネルに入ってください。」

ワンドロ企画で司会進行担当の人

## ぴよぽけワンドロとは

説明しよう。  
Discordで出されたお題を一時間で描く企画である！

参加方法は [@kemo_nano](https://twitter.com/kemo_nano) に聞いてください

## 機能

* ランダムでお題出し
* 開始、終了の時報
* みんながtwitterに上げた画像をdiscordに集める

## winでコンパイルする時

RasPI 無印だと6でいいはず。
PasPI 他のだと7になる。

```
set GOARCH=arm
set GOOS=linux
set GOARM=6
go build main.go
```

## ラズパイで実行する時

それぞれ

ex.
```
/xxxx/piyopoke/main -t odai
/xxxx/piyopoke/main -t before
/xxxx/piyopoke/main -t after
/xxxx/piyopoke/main -t watch
```

## 設定方法

crontab に設定する

ex.
```
50 21,23 * * * /xxxx/piyopoke/main -t odai -c /xxxx/piyopoke/config.json
00 22,00 * * * /xxxx/piyopoke/main -t before -c /xxxx/piyopoke/config.json
00 23,01 * * * /xxxx/piyopoke/main -t after -c /xxxx/piyopoke/config.json
00 22 * * * /xxxx/piyopoke/main -t watch -c /xxxx/piyopoke/config.json
```
