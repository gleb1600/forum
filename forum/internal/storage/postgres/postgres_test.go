package pgstorage_test

import (
	"context"
	"strings"

	"forum/internal/graph/model"
	"forum/internal/storage"
	pgstorage "forum/internal/storage/postgres"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/jackc/pgx/v5"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func newPost(id, title string, commentsEnabled bool) *model.Post {
	return &model.Post{
		ID:              id,
		Title:           title,
		Content:         "content",
		Author:          "author",
		CommentsEnabled: commentsEnabled,
	}
}

func newComment(id, postID, content string) *model.Comment {
	return &model.Comment{
		ID:      id,
		PostID:  postID,
		Author:  "author",
		Content: content,
	}
}

var (
	store       storage.Storage
	ctx         context.Context
	pgContainer *postgres.PostgresContainer
	connStr     string
)

var _ = BeforeSuite(func() {
	ctx = context.Background()
	var err error
	pgContainer, err = postgres.Run(ctx,
		"postgres:17",
		postgres.WithInitScripts("../../../docker-entrypoint-initdb.d/init.sql"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("user"),
		postgres.WithPassword("pass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
		),
	)
	Expect(err).NotTo(HaveOccurred())
	connStr, err = pgContainer.ConnectionString(ctx, "sslmode=disable")
	Expect(err).NotTo(HaveOccurred())
	store = pgstorage.NewPostgresStorage(connStr)
})

var _ = AfterSuite(func() {
	pgContainer.Terminate(ctx)
})

var _ = Describe("PostgresStorage", func() {
	BeforeEach(func() {
		conn, err := pgx.Connect(ctx, connStr)
		Expect(err).NotTo(HaveOccurred())
		defer conn.Close(ctx)
		_, err = conn.Exec(ctx, "TRUNCATE posts, comments")
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("ListPosts", func() {
		It("возвращает пустой список когда постов нет", func() {
			posts, err := store.ListPosts(ctx, 10, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(posts).To(BeEmpty())
		})

		It("возвращает не больше limit постов", func() {
			for i := 0; i < 5; i++ {
				Expect(store.CreatePost(ctx, newPost(
					strings.Repeat("post-id-", i+1),
					"title",
					true,
				))).To(Succeed())
			}

			posts, err := store.ListPosts(ctx, 3, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(posts).To(HaveLen(3))
		})
	})

	Describe("CreatePost и GetPost", func() {
		It("создаёт и возвращает пост по ID", func() {
			post := newPost("post-id-5", "title-1", true)
			Expect(store.CreatePost(ctx, post)).To(Succeed())

			got, err := store.GetPost(ctx, "post-id-5")
			Expect(err).NotTo(HaveOccurred())
			Expect(got.Title).To(Equal("title-1"))
		})

		It("возвращает ошибку для несуществующего поста", func() {
			_, err := store.GetPost(ctx, "nonexistent")
			Expect(err).To(MatchError(pgstorage.ErrPostNotFound))
		})
	})

	Describe("CreateComment", func() {
		BeforeEach(func() {
			Expect(store.CreatePost(ctx, newPost("post-id-5", "title", true))).To(Succeed())
		})

		It("создаёт комментарий к существующему посту, для первого комментария parent_id равен ID поста", func() {
			post := newPost("post-id-7", "title-1", true)
			Expect(store.CreatePost(ctx, post)).To(Succeed())

			comment := newComment("comm-id-1", "post-id-7", "comment-1")
			Expect(store.CreateComment(ctx, comment)).To(Succeed())

			comments, _ := store.GetCommentsByPost(ctx, "post-id-7", 1, 0)
			Expect(*comments[0].ParentID).To(Equal("post-id-7"))
		})

		It("для второго комментария parent_id равен ID предыдущего", func() {
			c1 := newComment("comm-id-2", "post-id-5", "comment-1")
			Expect(store.CreateComment(ctx, c1)).To(Succeed())

			c2 := newComment("comm-id-3", "post-id-5", "comment-2")
			Expect(store.CreateComment(ctx, c2)).To(Succeed())

			comments, _ := store.GetCommentsByPost(ctx, "post-id-5", 1, 1)
			Expect(*comments[0].ParentID).To(Equal("comm-id-2"))
		})

		It("возвращает ошибку если пост не найден", func() {
			Expect(store.CreateComment(ctx, newComment("comm-id-4", "unknown", "comment-1"))).
				To(MatchError(pgstorage.ErrPostNotFound))
		})

		It("возвращает ошибку если комментарии отключены", func() {
			Expect(store.CreatePost(ctx, newPost("post-id-8", "title", false))).To(Succeed())
			Expect(store.CreateComment(ctx, newComment("comm-id-5", "post-id-8", "comment-1"))).
				To(MatchError(pgstorage.ErrCommentsDisabled))
		})

		It("возвращает ошибку если контент длиннее 2000 символов", func() {
			long := strings.Repeat("а", 2001)
			Expect(store.CreateComment(ctx, newComment("comm-id-6", "post-id-5", long))).
				To(MatchError(pgstorage.ErrCommentTooLong))
		})
	})

	Describe("GetCommentsByPost", func() {
		BeforeEach(func() {
			Expect(store.CreatePost(ctx, newPost("post-id-9", "title", true))).To(Succeed())
			for i := 0; i < 4; i++ {
				Expect(store.CreateComment(ctx, newComment(
					strings.Repeat("comm9-id-", i+1),
					"post-id-9",
					"comment",
				))).To(Succeed())
			}
		})

		It("возвращает комментарии для поста", func() {
			comments, err := store.GetCommentsByPost(ctx, "post-id-9", 10, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(comments).NotTo(BeEmpty())
			Expect(len(comments)).To(BeNumerically("==", 4))
		})

		It("не возвращает больше чем limit", func() {
			comments, err := store.GetCommentsByPost(ctx, "post-id-9", 2, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(comments)).To(BeNumerically("==", 2))
		})

		It("возвращает ошибку для несуществующего поста", func() {
			_, err := store.GetCommentsByPost(ctx, "unknown", 10, 0)
			Expect(err).To(MatchError(pgstorage.ErrPostNotFound))
		})
	})

	Describe("DisableComments", func() {
		BeforeEach(func() {
			Expect(store.CreatePost(ctx, newPost("post-id-10", "title", true))).To(Succeed())
		})

		It("отключает комментарии у поста", func() {
			Expect(store.DisableComments(ctx, "post-id-10")).To(Succeed())

			post, err := store.GetPost(ctx, "post-id-10")
			Expect(err).NotTo(HaveOccurred())
			Expect(post.CommentsEnabled).To(BeFalse())
		})

		It("возвращает ошибку для несуществующего поста", func() {
			Expect(store.DisableComments(ctx, "unknown")).
				To(MatchError(pgstorage.ErrPostNotFound))
		})
	})
})
