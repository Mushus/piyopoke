# ぴよポケbot

ワンドロ企画で司会進行担当の人

##winでコンパイルする時

set PIYOPOKE_WH=https://discordapp.com/api/webhooks/389353176748392451/Tvs76XNf9FE_OfvCz3tw4Kr_rFWLu6LAPOHKBKByeFcR4NL3o_ZFqOzEidXXZEoAabrn
set GOARCH=arm
set GOOS=linux
set GOARM=6
go build main.go

## ラズパイで実行する時

それぞれ
```
PIYOPOKE_WH=xxx /xxx/xxx/main -t odai
PIYOPOKE_WH=xxx /xxx/xxx/main -t before
PIYOPOKE_WH=xxx /xxx/xxx/main -t after
```

## 設定方法

50 21,23 * * * PIYOPOKE_WH=xxx /xxx/xxx/main -t odai
0 22,0 * * * PIYOPOKE_WH=xxx /xxx/xxx/main -t before
0 23,1 * * * PIYOPOKE_WH=xxx /xxx/xxx/main -t after
