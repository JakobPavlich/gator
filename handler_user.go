package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/JakobPavlich/gator/internal/config"
	"github.com/JakobPavlich/gator/internal/database"
	"github.com/JakobPavlich/gator/internal/rss"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		fmt.Println("username required")
		os.Exit(1)
	}
	userName := cmd.arguments[0]
	_, err := s.db.GetUser(context.Background(), userName)
	if err != nil {
		fmt.Printf("User '%s' is not yet registered. Use the 'register (username)' command first.", userName)
		os.Exit(1)
	}
	cfg, err := config.Read()
	if err != nil {
		return errors.New("error retrieving config file")
	}
	cfg.SetUser(userName)
	s.cfg.CurrentUserName = cfg.CurrentUserName
	s.cfg.DBURL = cfg.DBURL
	fmt.Printf("User '%s' has been set \n", cfg.CurrentUserName)

	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.arguments) == 0 {
		fmt.Println("username required")
		os.Exit(1)
	}
	userName := cmd.arguments[0]

	userParams := database.CreateUserParams{}
	userParams.ID = uuid.New()
	userParams.CreatedAt = time.Now()
	userParams.UpdatedAt = time.Now()
	user, err := s.db.GetUser(context.Background(), userName)
	if err == nil {
		fmt.Printf("User '%s' already exists", user.Name)
		os.Exit(1)
	}
	userParams.Name = userName
	user, err = s.db.CreateUser(context.Background(), userParams)
	if err != nil {
		return fmt.Errorf("couldn't create a user: %v", err)
	}
	s.cfg.SetUser(user.Name)
	fmt.Println("User was successfully created.")
	fmt.Printf("Name: %s\n", user.Name)
	fmt.Printf("Data: %+v", user)

	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.DropTable(context.Background())
	if err != nil {
		return err
	}
	fmt.Println("Database was reset successfully")
	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}
	for _, user := range users {
		if user.Name == s.cfg.CurrentUserName {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}
	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.arguments) < 1 {
		return errors.New("needs argument time_between_reqs(1s, 1m, 1h,...)")
	}
	time_between_reqs, err := time.ParseDuration(cmd.arguments[0])
	if err != nil {
		return err
	}

	fmt.Printf("Collecting feeds every %v\n", cmd.arguments[0])

	ticker := time.NewTicker(time_between_reqs)
	for ; ; <-ticker.C {
		err := scrapeFeeds(s)
		if err != nil {
			return err
		}
	}
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	userId := user.ID
	if len(cmd.arguments) < 2 {
		return errors.New("command 'addfeed' needs two arguments (name,url)")
	}
	name := cmd.arguments[0]
	url := cmd.arguments[1]

	var params database.CreateFeedParams
	params.ID = uuid.New()
	params.Name = name
	params.CreatedAt = time.Now()
	params.UpdatedAt = time.Now()
	params.Url = url
	params.UserID = userId

	feed, err := s.db.CreateFeed(context.Background(), params)
	if err != nil {
		return fmt.Errorf("error entering the feed into the database: %v", err)
	}
	var paramsFeedFollow database.CreateFeedFollowParams
	paramsFeedFollow.ID = uuid.New()
	paramsFeedFollow.UserID = userId
	paramsFeedFollow.CreatedAt = time.Now()
	paramsFeedFollow.UpdatedAt = time.Now()
	paramsFeedFollow.FeedID = params.ID

	_, err = s.db.CreateFeedFollow(context.Background(), paramsFeedFollow)
	if err != nil {
		return fmt.Errorf("error creating feed: %v", err)
	}
	fmt.Println("New feed successfully added to the database")
	fmt.Printf("%+v", feed)
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	if len(cmd.arguments) > 0 {
		fmt.Println("'feeds' takes no arguments, it only lists the content:")
	}

	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("error retrieving feeds: %v", err)
	}
	fmt.Println("Printing feeds:")
	for id, feed := range feeds {
		name := feed.Name
		url := feed.Url
		userName, err := s.db.FindUserByID(context.Background(), feed.UserID)
		if err != nil {
			return fmt.Errorf("error retrieving user by ID: %v", err)
		}
		fmt.Printf("* Feed %d\n", id+1)
		fmt.Printf("  - Name: %s\n", name)
		fmt.Printf("  - Url: %s\n", url)
		fmt.Printf("  - User Name: %s\n", userName)

	}

	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) < 1 {
		return errors.New("missing url")
	}

	url := cmd.arguments[0]
	feed, err := s.db.FindFeedByURL(context.Background(), url)
	if err != nil {
		return errors.New("could not find a feed with the given URL")
	}

	params := database.CreateFeedFollowParams{}
	params.ID = uuid.New()
	params.CreatedAt = time.Now()
	params.UpdatedAt = time.Now()
	params.FeedID = feed.ID
	params.UserID = user.ID

	feedFollow, err := s.db.CreateFeedFollow(context.Background(), params)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" {
				// unique_violation - just skip this post
				return nil
			}
			return fmt.Errorf("error creating feed follow: %v", err)
		}
	}
	fmt.Printf("Feed: %s\n", feedFollow.FeedName)
	fmt.Printf("User: %s\n", feedFollow.UserName)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) > 0 {
		fmt.Println("No arguments possible for this command")
		fmt.Println()
	}
	feeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("error retrieving feeds for user: %v", err)
	}

	fmt.Printf("Printing feeds followed by user '%s':\n", user.Name)
	for _, feed := range feeds {
		fmt.Printf(" * %s\n", feed.FeedName)
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) < 1 {
		return errors.New("feed url needed")
	}
	feedURL := cmd.arguments[0]
	feed, err := s.db.FindFeedByURL(context.Background(), feedURL)
	if err != nil {
		return err
	}

	var params database.DeleteFeedFollowParams
	params.UserID = user.ID
	params.FeedID = feed.ID

	err = s.db.DeleteFeedFollow(context.Background(), params)
	if err != nil {
		return err
	}
	fmt.Printf("Unfollowed '%s'\n", feed.Name)
	return nil
}

func scrapeFeeds(s *state) error {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}
	err = s.db.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		return err
	}
	fetchedFeed, err := rss.FetchFeed(context.Background(), feed.Url)
	if err != nil {
		return err
	}
	for _, item := range fetchedFeed.Channel.Item {
		title := html.UnescapeString(item.Title)
		description := html.UnescapeString(item.Description)
		pubDate, err := time.Parse(time.RFC1123, item.PubDate)
		if err != nil {
			return err
		}
		sqlPubDate := sql.NullTime{Time: pubDate, Valid: true}
		sqlDescription := sql.NullString{String: description, Valid: true}

		var params database.CreatePostParams
		params.ID = uuid.New()
		params.CreatedAt = time.Now()
		params.UpdatedAt = time.Now()
		params.Title = title
		params.PublishedAt = sqlPubDate
		params.Description = sqlDescription
		params.Url = item.Link
		params.FeedID = feed.ID

		_, err = s.db.CreatePost(context.Background(), params)
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok {
				if pqErr.Code == "23505" {
					// unique_violation - just skip this post
					return nil
				}
				return fmt.Errorf("error creating feed follow: %v", err)
			}
		}
		// if err != nil {
		// 	return err
		// }
		// fmt.Printf(" - saved post '%s'\n", post.Title)
	}

	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	limit := 2
	if len(cmd.arguments) > 0 {
		argLimit, err := strconv.Atoi(cmd.arguments[0])
		if err != nil {
			return err
		}
		limit = argLimit
	}
	var params database.GetPostsForUserParams
	feedFollows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return err
	}
	fmt.Printf("Printing posts for user '%s':\n", user.Name)
	for _, f := range feedFollows {
		params.FeedID = f.FeedID
		params.Limit = int32(limit)
		posts, err := s.db.GetPostsForUser(context.Background(), params)
		if err != nil {
			return err
		}
		fmt.Printf("Feed: %s\n", f.Name)
		for _, post := range posts {
			fmt.Printf(" * %s\n", strings.ToUpper(post.Title))
			fmt.Println(post.Description.String)
		}
		fmt.Println()
	}
	return nil
}
