package mmrstorage

import (
	"context"
	"errors"
	"forum/internal/graph/model"
	"forum/internal/storage"
	"sync"
)

var (
	ErrNotFound              = errors.New("not found")
	ErrCommentsDisabled      = errors.New("comments are disabled")
	ErrPostNotFound          = errors.New("post not found")
	ErrCommentTooLong        = errors.New("comments are too long")
	ErrParentCommentNotFound = errors.New("comment's parent not found")
)

type memoryStorage struct {
	posts          map[string]*model.Post
	postsMutex     sync.RWMutex
	comments       map[string]*model.Comment
	commentsMutex  sync.RWMutex
	commentsMutex2 sync.RWMutex
	commentSubs    map[string]map[chan *model.Comment]struct{}
}

func NewMemoryStorage() storage.Storage {
	return &memoryStorage{
		posts:       make(map[string]*model.Post),
		comments:    make(map[string]*model.Comment),
		commentSubs: make(map[string]map[chan *model.Comment]struct{}),
	}
}

func (m *memoryStorage) CreatePost(ctx context.Context, post *model.Post) error {
	m.postsMutex.Lock()
	defer m.postsMutex.Unlock()

	m.posts[post.ID] = post
	return nil
}
func (m *memoryStorage) GetPost(ctx context.Context, id string) (*model.Post, error) {
	post := m.posts[id]
	if post == nil {
		return nil, ErrPostNotFound
	}
	return post, nil
}
func (m *memoryStorage) ListPosts(ctx context.Context, limit, offset int) ([]*model.Post, error) {
	posts := make([]*model.Post, 0)
	for _, v := range m.posts {
		posts = append(posts, v)
	}
	return posts, nil
}
func (m *memoryStorage) CreateComment(ctx context.Context, comment *model.Comment) error {
	m.commentsMutex.Lock()
	defer m.commentsMutex.Unlock()

	post := m.posts[comment.PostID]
	if post == nil {
		return ErrPostNotFound
	}
	if !post.CommentsEnabled {
		return ErrCommentsDisabled
	}
	if len([]rune(comment.Content)) > 2000 {
		return ErrCommentTooLong
	}
	comments, err := m.GetCommentsByPost(ctx, comment.PostID, 0, 0)
	if err != nil {
		return err
	}
	if len(comments) == 0 {
		comment.ParentID = &comment.PostID
	} else {
		comment.ParentID = &comments[len(comments)-1].ID
	}
	m.comments[comment.ID] = comment
	return nil
}
func (m *memoryStorage) GetCommentsByPost(ctx context.Context, postID string, limit, offset int) ([]*model.Comment, error) {
	m.commentsMutex2.Lock()
	defer m.commentsMutex2.Unlock()
	post := m.posts[postID]
	if post == nil {
		return nil, ErrPostNotFound
	}
	comments := make([]*model.Comment, 0)
	for _, v := range m.comments {
		if v.PostID == postID {
			comments = append(comments, v)
		}
	}
	return comments, nil
}
func (m *memoryStorage) DisableComments(ctx context.Context, postID string) error {
	post := m.posts[postID]
	if post == nil {
		return ErrPostNotFound
	}
	m.posts[postID].CommentsEnabled = false
	return nil
}
