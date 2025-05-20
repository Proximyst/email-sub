package models

import "time"

type FeedPostEmailRequest struct {
	// Feed is the URL to the feed.
	Feed string `json:"feed"`
	// ID is the ID of the post, as defined by the feed.
	ID string `json:"id"`
	// Email is the email address of the user.
	Email string `json:"email"`
	// Link is the URL to the post. It is used to send the email.
	Link string `json:"link"`
	// Posted is when the post was published. It is used to determine the order of posts.
	Posted time.Time `json:"posted"`
}
