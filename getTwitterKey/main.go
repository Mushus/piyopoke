package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/dghubble/oauth1"
	"github.com/dghubble/oauth1/twitter"
)

var config = oauth1.Config{
	ConsumerKey:    "",
	ConsumerSecret: "",
	CallbackURL:    "",
	Endpoint:       twitter.AuthorizeEndpoint,
}

func main() {
	requestToken, _, _ := config.RequestToken()
	authorizationURL, _ := config.AuthorizationURL(requestToken)
	fmt.Println(authorizationURL)
	fmt.Print("pin? ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	verifier := scanner.Text()
	accessToken, accessSecret, _ := config.AccessToken(requestToken, "", verifier)
	fmt.Printf("access_token: %s\n", accessToken)
	fmt.Printf("access_secret: %s\n", accessSecret)
}
