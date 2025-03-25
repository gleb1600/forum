// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"time"
)

type Comment struct {
	ID        string    `json:"id"`
	PostID    string    `json:"postId"`
	ParentID  *string   `json:"parentId,omitempty"`
	Author    string    `json:"author"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

type Mutation struct {
}

type NewComment struct {
	PostID   string  `json:"postId"`
	ParentID *string `json:"parentId,omitempty"`
	Author   string  `json:"author"`
	Content  string  `json:"content"`
}

type NewPost struct {
	Title           string `json:"title"`
	Content         string `json:"content"`
	Author          string `json:"author"`
	CommentsEnabled bool   `json:"commentsEnabled"`
}

type Post struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	Content         string    `json:"content"`
	Author          string    `json:"author"`
	CommentsEnabled bool      `json:"commentsEnabled"`
	CreatedAt       time.Time `json:"createdAt"`
}

type PostWithComments struct {
	Post     *Post      `json:"post"`
	Comments []*Comment `json:"comments"`
}

type Query struct {
}

type Subscription struct {
}
