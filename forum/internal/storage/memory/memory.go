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
	posts         map[string]*model.Post
	postsMutex    sync.RWMutex
	comments      map[string]*model.Comment
	commentsMutex sync.RWMutex
	commentSubs   map[string]map[chan *model.Comment]struct{}
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
	m.postsMutex.RLock()
	defer m.postsMutex.RUnlock()

	post := m.posts[id]
	if post == nil {
		return nil, ErrPostNotFound
	}
	return post, nil
}
func (m *memoryStorage) ListPosts(ctx context.Context, limit, offset int) ([]*model.Post, error) {
	posts := make([]*model.Post, 0, limit)

	m.postsMutex.RLock()
	for _, v := range m.posts {
		if limit < 1 {
			break
		}

		posts = append(posts, v)
		limit--
	}
	m.postsMutex.RUnlock()

	return posts, nil
}
func (m *memoryStorage) CreateComment(ctx context.Context, comment *model.Comment) error {
	m.postsMutex.RLock()

	post := m.posts[comment.PostID]
	m.postsMutex.RUnlock()

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

	m.commentsMutex.Lock()
	m.comments[comment.ID] = comment
	m.commentsMutex.Unlock()
	return nil
}
func (m *memoryStorage) GetCommentsByPost(ctx context.Context, postID string, limit, offset int) ([]*model.Comment, error) {
	m.postsMutex.RLock()
	post := m.posts[postID]
	m.postsMutex.RUnlock()

	if post == nil {
		return nil, ErrPostNotFound
	}
	comments := make([]*model.Comment, 0)
	m.commentsMutex.RLock()

	for _, v := range m.comments {
		if limit < 1 {
			break
		}
		if v.PostID == postID {
			comments = append(comments, v)
			limit--
		}
	}

	m.commentsMutex.RUnlock()

	return comments, nil
}
func (m *memoryStorage) DisableComments(ctx context.Context, postID string) error {
	m.postsMutex.Lock()
	defer m.postsMutex.Unlock()

	post := m.posts[postID]
	if post == nil {
		return ErrPostNotFound
	}
	m.posts[postID].CommentsEnabled = false
	return nil
}
