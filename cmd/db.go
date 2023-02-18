package main

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"time"
)

type Post struct {
	Id                    string
	FullId                string `gorm:"primaryKey"`
	Created               time.Time
	Edited                time.Time
	Permalink             string
	Url                   string
	Title                 string
	Body                  string
	Likes                 *bool
	Score                 int
	UpvoteRatio           float32
	NumberOfComments      int
	SubredditName         string
	SubredditNamePrefixed string
	SubredditId           string
	SubredditSubscribers  int
	Author                string
	AuthorId              string
	Spoiler               bool
	Locked                bool
	Nsfw                  bool
	IsSelfPost            bool
	Saved                 bool
	Stickied              bool
}

type Subreddit struct {
	Name              string `gorm:"primaryKey"`
	LastProcessedPage string
}

var DB *gorm.DB

func OpenDB() {
	fmt.Println("Opening DB")
	if DB != nil {
		return
	}
	var err error
	DB, err = gorm.Open(sqlite.Open("subreddits.sqlite3"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	err = DB.AutoMigrate(&Post{})
	if err != nil {
		panic(err)
	}
	err = DB.AutoMigrate(&Subreddit{})
	if err != nil {
		panic(err)
	}
	fmt.Println("DB Opened")
}
