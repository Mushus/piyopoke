package main

import (
	"strings"
	"testing"
	"time"
)

func setNow(t time.Time) {
	now = func() time.Time {
		return t
	}
}

func TestTSV(t *testing.T) {
	date, _ := time.Parse(time.RFC822, "02 Jan 06 15:04 MST")
	setNow(date)
	r := strings.NewReader("01/02\tポケモン")
	chara := findToTSV(r)
	if chara != "ポケモン" {
		t.Fatalf("chara is not ポケモン")
	}
}

func TestTSV2(t *testing.T) {
	date, _ := time.Parse(time.RFC822, "02 Jan 06 15:04 MST")
	setNow(date)
	r := strings.NewReader("01/01\tポケモン")
	chara := findToTSV(r)
	if chara == "ポケモン" {
		t.Fatalf("chara is ポケモン")
	}
}
