package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	"errors"

	"github.com/google/uuid"
	"github.com/jonasyke/gator/internal/database"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("the login handler expects a username")
	}

	username := cmd.args[0]

	_, err := s.db.GetUser(context.Background(), username)
	if err != nil {
		return fmt.Errorf("user %s does not exist: %w", username, err)
	}

	err = s.cfg.SetUser(username)
	if err != nil {
		return fmt.Errorf("could not set current user: %w", err)
	}

	fmt.Printf("User has been set to: %s\n", username)
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("the register handler expects a username")
	}

	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
	})
	if err != nil {
		return fmt.Errorf("could not create user: %w", err)
	}

	err = s.cfg.SetUser(user.Name)
	if err != nil {
		return fmt.Errorf("could not set current user: %w", err)
	}

	fmt.Printf("User created successfully: %v\n", user)
	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.Reset(context.Background())
	if err != nil {
		return fmt.Errorf("could not reset users: %w", err)
	}

	fmt.Println("All users have been deleted.")
	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("could not get users: %w", err)
	}

	for _, user := range users {
		suffix := ""
		if user.Name == s.cfg.Username {
			suffix = " (current)"
		}
		fmt.Printf("* %s%s\n", user.Name, suffix)
	}
	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: %s", cmd.name)
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return fmt.Errorf("invalid duration: %w", err)
	}

	fmt.Printf("Collecting feeds every %s...\n", timeBetweenRequests)

	ticker := time.NewTicker(timeBetweenRequests)
	for ; ; <-ticker.C {
		err := scrapeFeeds(s)
		if err != nil {
			fmt.Printf("Error scraping feeds: %v\n", err)
		}
	}
}

func handlerAddFeed(s *state, cmd command, currentUser database.User) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("usage: %s <name> <url>", cmd.name)
	}

	newFeed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      cmd.args[0],
		Url:       cmd.args[1],
		UserID:    currentUser.ID,
	})
	if err != nil {
		return fmt.Errorf("could not create new feed: %w", err)
	}
	follow, err := createFeedFollow(context.Background(), s.db, currentUser.ID, newFeed.ID)
	if err != nil {
		return fmt.Errorf("could not create new follow: %w", err)
	}

	fmt.Printf("Feed name: %s\n", follow.FeedName)
	fmt.Printf("User: %s\n", follow.UserName)

	return nil
}

func handlerFeeds(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("%s command takes no arguments", cmd.name)
	}

	allFeeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("unable to retrieve feeds: %s", err)
	}

	for _, feed := range allFeeds {
		fmt.Printf("* Name: %s\n", feed.FeedName)
		fmt.Printf("  URL: %s\n", feed.Url)
		fmt.Printf("  UserName: %s\n", feed.UserName)
	}
	return nil
}

func handlerFollow(s *state, cmd command, currentUser database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("%s only takes a URL", cmd.name)
	}

	url := cmd.args[0]

	feed, err := s.db.GetFeedByUrl(context.Background(), url)
	if err != nil {
		return fmt.Errorf("could not find feed with url: %s: %w", url, err)
	}

	new_follow, err := createFeedFollow(context.Background(), s.db, currentUser.ID, feed.ID)

	if err != nil {
		return fmt.Errorf("could not generate follow: %w", err)
	}
	fmt.Printf("Feed name: %s\n", new_follow.FeedName)
	fmt.Printf("current user: %s\n", new_follow.UserName)
	return nil

}

func handlerFollowing(s *state, cmd command, currentUser database.User) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("%s does not take and argument", cmd.name)
	}

	feeds, err := s.db.GetFeedFollowsByUser(context.Background(), currentUser.ID)
	if err != nil {
		return fmt.Errorf("could not retrieve feeds %w", err)
	}

	for _, feed := range feeds {
		fmt.Println(feed.FeedName)
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, currentUser database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("needs the url only")
	}
	url := cmd.args[0]

	feed, err := s.db.GetFeedByUrl(context.Background(), url)
	if err != nil {
		return fmt.Errorf("could not retrieve feed: %s :%w", url, err)
	}

	err = s.db.FeedUnfollow(context.Background(), database.FeedUnfollowParams{
		UserID: currentUser.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("could not perform unfollow: %w", err)
	}
	return nil
}

func createFeedFollow(ctx context.Context, db *database.Queries, userID uuid.UUID, feedID uuid.UUID) (database.CreateFeedFollowRow, error) {
	return db.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    userID,
		FeedID:    feedID,
	})
}

func handlerBrowse(s *state, cmd command) error {
	limit := 2
	if len(cmd.args) > 0 {
		if l, err := strconv.Atoi(cmd.args[0]); err == nil {
			limit = l
		}
	}

	user, err := s.db.GetUser(context.Background(), s.cfg.Username)
	if err != nil {
		return fmt.Errorf("Could not fetch user in browse: %w", err)
	}

	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})
	for _, post := range posts {
		fmt.Printf("--- %s ---\n", post.Title.String)
		fmt.Printf("Source: %s\n", post.Url)
		fmt.Printf("Published: %v\n", post.PublishedAt.Time.Format("Jan 02, 2006"))
		fmt.Printf("Description: %s\n\n", post.Description.String)
	}
	return nil
}

func scrapeFeeds(s *state) error {
	nextFeed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("could not retrieve the next feed: %w", err)
	}

	markedFeed, err := s.db.MarkFeedFetched(context.Background(), nextFeed.ID)
	if err != nil {
		return fmt.Errorf("could not mark feed fetched: %w", err)
	}

	newFeed, err := FetchFeed(context.Background(), markedFeed.Url)
	if err != nil {
		return fmt.Errorf("could not fetch new feed: %w", err)
	}

	for _, item := range newFeed.Channel.Items {
		pubAt, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			log.Printf("couldn't parse date %v: %v", item.PubDate, err)
			continue
		}
		_, err = s.db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Title:       sql.NullString{String: item.Title, Valid: true},
			Url:         item.Link,
			Description: sql.NullString{String: item.Description, Valid: true},
			PublishedAt: sql.NullTime{Time: pubAt, Valid: true},
			FeedID:      nextFeed.ID,
		})
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				continue
			}

			if errors.Is(err, sql.ErrNoRows) {
				continue
			}

			log.Printf("could not create new post: %v", err)
			continue
		}

	}
	return nil
}
