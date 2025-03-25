package graph

import (
	"forum/internal/graph/model"
	"forum/internal/storage"
	"sync"
)

type Resolver struct {
	ResolverStorage storage.Storage
	SubStorage      map[[2]string]chan *model.Comment
	mu              sync.Mutex
}
