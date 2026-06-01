package mmrstorage_test

import (
	"context"
	"strings"

	"forum/internal/graph/model"
	"forum/internal/storage"
	mmrstorage "forum/internal/storage/memory"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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

var _ = Describe("MemoryStorage", func() {
	var (
		store storage.Storage
		ctx   context.Context
	)

	BeforeEach(func() {
		store = mmrstorage.NewMemoryStorage()
		ctx = context.Background()
	})

	Describe("CreatePost и GetPost", func() {
		It("создаёт и возвращает пост по ID", func() {
			post := newPost("post-id-1", "title-1", true)
			Expect(store.CreatePost(ctx, post)).To(Succeed())

			got, err := store.GetPost(ctx, "post-id-1")
			Expect(err).NotTo(HaveOccurred())
			Expect(got.Title).To(Equal("title-1"))
		})

		It("возвращает ошибку для несуществующего поста", func() {
			_, err := store.GetPost(ctx, "nonexistent")
			Expect(err).To(MatchError(mmrstorage.ErrPostNotFound))
		})
	})

	Describe("ListPosts", func() {
		BeforeEach(func() {
			for i := 0; i < 5; i++ {
				Expect(store.CreatePost(ctx, newPost(
					strings.Repeat("post-id-", i+1),
					"title",
					true,
				))).To(Succeed())
			}
		})

		It("возвращает не больше limit постов", func() {
			posts, err := store.ListPosts(ctx, 3, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(posts).To(HaveLen(3))
		})

		It("возвращает пустой список когда постов нет", func() {
			emptyStore := mmrstorage.NewMemoryStorage()
			posts, err := emptyStore.ListPosts(ctx, 10, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(posts).To(BeEmpty())
		})
	})

	Describe("CreateComment", func() {
		BeforeEach(func() {
			Expect(store.CreatePost(ctx, newPost("post-id-1", "title", true))).To(Succeed())
		})

		It("создаёт комментарий к существующему посту, для первого комментария parent_id равен ID поста", func() {
			comment := newComment("comm-id-1", "post-id-1", "comment-1")
			Expect(store.CreateComment(ctx, comment)).To(Succeed())
			Expect(*comment.ParentID).To(Equal("post-id-1"))
		})

		It("для второго комментария parent_id равен ID предыдущего", func() {
			c1 := newComment("comm-id-1", "post-id-1", "comment-1")
			Expect(store.CreateComment(ctx, c1)).To(Succeed())

			c2 := newComment("comm-id-2", "post-id-1", "comment-2")
			Expect(store.CreateComment(ctx, c2)).To(Succeed())
			Expect(*c2.ParentID).To(Equal("comm-id-1"))
		})

		It("возвращает ошибку если пост не найден", func() {
			Expect(store.CreateComment(ctx, newComment("comm-id-1", "unknown", "comment-1"))).
				To(MatchError(mmrstorage.ErrPostNotFound))
		})

		It("возвращает ошибку если комментарии отключены", func() {
			Expect(store.CreatePost(ctx, newPost("post-id-1", "title", false))).To(Succeed())
			Expect(store.CreateComment(ctx, newComment("comm-id-1", "post-id-1", "comment-1"))).
				To(MatchError(mmrstorage.ErrCommentsDisabled))
		})

		It("возвращает ошибку если контент длиннее 2000 символов", func() {
			long := strings.Repeat("а", 2001)
			Expect(store.CreateComment(ctx, newComment("comm-id-1", "post-id-1", long))).
				To(MatchError(mmrstorage.ErrCommentTooLong))
		})
	})

	Describe("GetCommentsByPost", func() {
		BeforeEach(func() {
			Expect(store.CreatePost(ctx, newPost("post-id-1", "title", true))).To(Succeed())
			for i := 0; i < 4; i++ {
				Expect(store.CreateComment(ctx, newComment(
					strings.Repeat("comm-id-", i+1),
					"post-id-1",
					"comment",
				))).To(Succeed())
			}
		})

		It("возвращает комментарии для поста", func() {
			comments, err := store.GetCommentsByPost(ctx, "post-id-1", 10, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(comments).NotTo(BeEmpty())
		})

		It("не возвращает больше чем limit", func() {
			comments, err := store.GetCommentsByPost(ctx, "post-id-1", 2, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(comments)).To(BeNumerically("<=", 2))
		})

		It("возвращает ошибку для несуществующего поста", func() {
			_, err := store.GetCommentsByPost(ctx, "unknown", 10, 0)
			Expect(err).To(MatchError(mmrstorage.ErrPostNotFound))
		})
	})

	Describe("DisableComments", func() {
		BeforeEach(func() {
			Expect(store.CreatePost(ctx, newPost("post-id-1", "title", true))).To(Succeed())
		})

		It("отключает комментарии у поста", func() {
			Expect(store.DisableComments(ctx, "post-id-1")).To(Succeed())

			post, err := store.GetPost(ctx, "post-id-1")
			Expect(err).NotTo(HaveOccurred())
			Expect(post.CommentsEnabled).To(BeFalse())
		})

		It("возвращает ошибку для несуществующего поста", func() {
			Expect(store.DisableComments(ctx, "unknown")).
				To(MatchError(mmrstorage.ErrPostNotFound))
		})
	})
})
