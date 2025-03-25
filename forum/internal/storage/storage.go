package storage

import (
	"context"
	"forum/internal/graph/model"
)

type Storage interface {
	PostStore
	CommentStore
}

type PostStore interface {
	CreatePost(ctx context.Context, post *model.Post) error
	GetPost(ctx context.Context, id string) (*model.Post, error)
	ListPosts(ctx context.Context, limit, offset int) ([]*model.Post, error)
	DisableComments(ctx context.Context, postID string) error
}

type CommentStore interface {
	CreateComment(ctx context.Context, comment *model.Comment) error
	GetCommentsByPost(ctx context.Context, postID string, limit, offset int) ([]*model.Comment, error)
}

func NewSubStorage() map[[2]string]chan *model.Comment {
	return make(map[[2]string]chan *model.Comment)
}
