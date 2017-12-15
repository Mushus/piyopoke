# ぴよポケbot

ワンドロ企画で司会進行担当の人

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

```
PIYOPOKE_WH=xxx /xxx/xxx/main -t odai
PIYOPOKE_WH=xxx /xxx/xxx/main -t before
PIYOPOKE_WH=xxx /xxx/xxx/main -t after
```

## 設定方法

crontab に設定する
```
50 21,23 * * * PIYOPOKE_WH=xxx /xxx/xxx/main -t odai
00 22,00 * * * PIYOPOKE_WH=xxx /xxx/xxx/main -t before
00 23,01 * * * PIYOPOKE_WH=xxx /xxx/xxx/main -t after
00 22 * * * PIYOPOKE_WH=xxx PIYOPOKE_CONSUMER_KEY=xxx PIYOPOKE_CONSUMER_SECRET=xxx PIYOPOKE_ACCESS_TOKEN=xxx PIYOPOKE_ACCESS_SECRET=xxx /home/mushus/piyopoke/main -t watch

```
