package main

import (
	"context"
	"fmt"
	"github.com/vartanbeno/go-reddit/v2/reddit"
	"time"
)

var client *reddit.Client

func StartUpdater(subRedditName string) {
	// Check if we have a record of this subreddit
	var sr Subreddit
	DB.Where("name = ?", subRedditName).FirstOrCreate(&sr)
	if sr.LastProcessedPage == "" {
		UpdatePage(subRedditName, "", true)
	} else {
		UpdateFrontPage(subRedditName)
		UpdatePage(subRedditName, sr.LastProcessedPage, true)
	}
}

func UpdateFrontPage(name string) {
	UpdatePage(name, "", false)
}

func UpdatePage(subName, after string, loop bool) {
	var subreddit []*reddit.Post
	var r *reddit.Response
	var err error
	if after == "" {
		fmt.Println("Downloading first page")
		subreddit, r, err = client.Subreddit.NewPosts(context.Background(), subName, &reddit.ListOptions{
			Limit: 100,
		})
		if err != nil {
			panic(err)
		}
	} else {
		fmt.Printf("Downloading after %v\n", after)
		subreddit, r, err = client.Subreddit.NewPosts(context.Background(), subName, &reddit.ListOptions{
			Limit: 100,
			After: after,
		})
		if err != nil {
			panic(err)
		}
	}

	tx := DB.Begin()

	for _, post := range subreddit {
		// Check if we have this post already (by FullId)
		var p int64
		err = tx.Where("full_id = ?", post.FullID).Count(&p).Error
		if err == nil || p > 0 {
			fmt.Println("Skipping post (already exists)", post.FullID)
			// We have this post already, skip it
			continue
		}

		err = tx.Create(&Post{
			Id:                    post.ID,
			FullId:                post.FullID,
			Created:               post.Created.Time,
			Edited:                post.Edited.Time,
			Permalink:             post.Permalink,
			Url:                   post.URL,
			Title:                 post.Title,
			Body:                  post.Body,
			Likes:                 post.Likes,
			Score:                 post.Score,
			UpvoteRatio:           post.UpvoteRatio,
			NumberOfComments:      post.NumberOfComments,
			SubredditName:         post.SubredditName,
			SubredditNamePrefixed: post.SubredditNamePrefixed,
			SubredditId:           post.SubredditID,
			SubredditSubscribers:  post.SubredditSubscribers,
			Author:                post.Author,
			AuthorId:              post.AuthorID,
			Spoiler:               post.Spoiler,
			Locked:                post.Locked,
			Nsfw:                  post.NSFW,
			IsSelfPost:            post.IsSelfPost,
			Saved:                 post.Saved,
			Stickied:              post.Stickied,
		}).Error
		if err != nil {
			// If UNIQUE constraint failed: posts.full_id ignore it
			if err.Error() != "UNIQUE constraint failed: posts.full_id" {
				panic(err)
			}
		}
	}

	if r.Rate.Remaining <= r.Rate.Used {
		fmt.Printf("Sleeping for %v seconds\n", r.Rate.Reset.Sub(time.Now()).Seconds())
		time.Sleep(r.Rate.Reset.Sub(time.Now()) + (5 * time.Second))
	} else {
		fmt.Printf("Rate limit remaining: %v\n", r.Rate.Remaining)
		fmt.Printf("Rate limit used: %v\n", r.Rate.Used)
		fmt.Printf("Rate limit reset: %v\n", r.Rate.Reset.Sub(time.Now()).Seconds())

		// Calculate how long we should sleep to not exceed the rate limit

		cyclesRemaining := r.Rate.Remaining / r.Rate.Used
		secondsRemaining := r.Rate.Reset.Sub(time.Now()).Seconds()
		secondsPerCycle := secondsRemaining / float64(cyclesRemaining)
		fmt.Printf("Sleeping for %v seconds\n", secondsPerCycle)
		//time.Sleep(time.Duration(secondsPerCycle) * time.Second)
	}
	if after != "" {
		var sr Subreddit
		tx.Where("name = ?", "battlemaps").FirstOrCreate(&sr)
		sr.Name = subName
		sr.LastProcessedPage = after
		tx.Save(&sr)
	}
	tx.Commit()
	if r.After != "" {
		if loop {
			UpdatePage(subName, r.After, loop)
		}
	} else {
		fmt.Printf("No more pages (%+v)\n", r)
	}
}
