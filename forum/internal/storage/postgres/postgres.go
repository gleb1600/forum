package pgstorage

import (
	"context"
	"errors"
	"log"
	"time"

	"forum/internal/graph/model"
	"forum/internal/storage"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound              = errors.New("not found")
	ErrCommentsDisabled      = errors.New("comments are disabled")
	ErrPostNotFound          = errors.New("post not found")
	ErrCommentTooLong        = errors.New("comment is too long")
	ErrParentCommentNotFound = errors.New("comment's parent not found")
)

type postgresStorage struct {
	pool *pgxpool.Pool
}

func NewPostgresStorage(connString string) storage.Storage {
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		log.Fatalf("Failed to initialize PostgreSQL storage")
		return nil
	}
	return &postgresStorage{pool: pool}
}

func (p *postgresStorage) CreatePost(ctx context.Context, post *model.Post) error {
	_, err := p.pool.Exec(ctx,
		`INSERT INTO posts (id, title, content, author, comments_enabled, created_at)
         VALUES ($1, $2, $3, $4, $5, $6)`,
		post.ID, post.Title, post.Content, post.Author, post.CommentsEnabled, time.Now())
	return err
}

func (p *postgresStorage) GetPost(ctx context.Context, id string) (*model.Post, error) {
	var post model.Post
	err := p.pool.QueryRow(ctx,
		`SELECT id, title, content, author, comments_enabled, created_at 
         FROM posts WHERE id = $1`, id).
		Scan(&post.ID, &post.Title, &post.Content, &post.Author, &post.CommentsEnabled, &post.CreatedAt)

	if err == pgx.ErrNoRows {
		return nil, ErrPostNotFound
	}
	return &post, err
}

func (p *postgresStorage) ListPosts(ctx context.Context, limit, offset int) ([]*model.Post, error) {
	rows, err := p.pool.Query(ctx,
		`SELECT id, title, content, author, comments_enabled, created_at 
         FROM posts ORDER BY created_at LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*model.Post
	for rows.Next() {
		var post model.Post
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Author, &post.CommentsEnabled, &post.CreatedAt)
		if err != nil {
			return nil, err
		}
		posts = append(posts, &post)
	}
	return posts, nil
}

func (p *postgresStorage) CreateComment(ctx context.Context, comment *model.Comment) error {
	var commentsEnabled bool
	err := p.pool.QueryRow(ctx,
		`SELECT comments_enabled FROM posts WHERE id = $1`, comment.PostID).
		Scan(&commentsEnabled)

	if err != nil {
		return ErrPostNotFound
	}

	if !commentsEnabled {
		return ErrCommentsDisabled
	}

	if len(comment.Content) > 2000 {
		return ErrCommentTooLong
	}

	err = p.pool.QueryRow(ctx,
		`SELECT id FROM comments WHERE post_id = $1`, comment.PostID).Scan()

	var idParent string
	if err == pgx.ErrNoRows {
		idParent = comment.PostID
	} else {
		var idLastComment string
		p.pool.QueryRow(ctx, `SELECT id FROM comments WHERE post_id = $1
		 ORDER BY created_at DESC LIMIT 1`, comment.PostID).Scan(&idLastComment)
		idParent = idLastComment
	}

	_, err = p.pool.Exec(ctx,
		`INSERT INTO comments (id, post_id, parent_id, content, author, created_at)
         VALUES ($1, $2, $3, $4, $5, $6)`,
		comment.ID, comment.PostID, idParent, comment.Content, comment.Author, time.Now())

	if err != nil {
		return err
	}

	return nil
}

func (p *postgresStorage) GetCommentsByPost(ctx context.Context, postID string, limit, offset int) ([]*model.Comment, error) {

	rows, err := p.pool.Query(ctx,
		`SELECT id, post_id, parent_id, content, author, created_at 
		FROM comments WHERE post_id = $1 ORDER BY created_at LIMIT $2 OFFSET $3`,
		postID, limit, offset)

	if err != nil {
		return nil, ErrPostNotFound
	}
	defer rows.Close()

	var comments []*model.Comment
	for rows.Next() {
		var comment model.Comment
		err := rows.Scan(&comment.ID, &comment.PostID, &comment.ParentID, &comment.Content, &comment.Author, &comment.CreatedAt)
		if err != nil {
			return nil, err
		}
		comments = append(comments, &comment)
	}
	return comments, nil
}

func (p *postgresStorage) DisableComments(ctx context.Context, postID string) error {
	_, err := p.pool.Exec(ctx,
		`UPDATE posts SET comments_enabled = $1 WHERE id = $2`,
		false, postID)
	if err != nil {
		return err
	}
	return nil
}
