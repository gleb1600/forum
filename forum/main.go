package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"forum/internal/graph"
	storage "forum/internal/storage"
	mmrstorage "forum/internal/storage/memory"
	pgstorage "forum/internal/storage/postgres"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/vektah/gqlparser/v2/ast"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	var storageType string
	var dsn string
	var store storage.Storage

	flag.StringVar(&storageType, "storage", "", "Storage type (memory|postgres)")
	flag.Parse()

	subStore := storage.NewSubStorage()

	switch storageType {
	case "memory":
		store = mmrstorage.NewMemoryStorage()
		log.Println("Using in-memory storage")
	case "postgres":
		dsn = "postgres://forum_user:secret@localhost:5431/forumdb?sslmode=disable"
		store = pgstorage.NewPostgresStorage(dsn)
		log.Println("Using PostgreSQL storage")
	default:
		log.Fatalf("Unknown storage type: %s", storageType)
		log.Fatalf("Storage type (memory|postgres)")
	}

	rslvr := &graph.Resolver{ResolverStorage: store, SubStorage: subStore}
	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: rslvr}))

	srv.AddTransport(transport.Websocket{})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
