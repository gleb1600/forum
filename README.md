# Система для управления постами и комментариями с использованием GraphQL и PostgreSQL. Аналогично функционалу Habr/Reddit.

## Особенности проекта
- Создание/просмотр постов с пагинацией
- Иерархические комментарии под постами
- Возможность разрешить/запретить комментирование поста
- Rеaltime-обновления через GraphQL Subscriptions
- Запуск через docker-compose, работа с образом postgres
- Поддержка двух хранилищ: In-Memory и PostgreSQL

## Работа

1. **Запуск**:
```bash
docker-compose up -d
```
2. **Запустите с флагом -storage**:

Для хранения данных в памяти:
```bash
go run main.go -storage memory
```
Для хранения данных в PostgreSQL:
```bash
go run main.go -storage postgres
```
3. **Перейдите на http://localhost:8080/**:

Для создания поста:
```graphql
mutation createPost {
    createPost(
        input: {
            title: "My first post",
            content: "Hi, my name is.. I like ...",
            author: "Nika",
            commentsEnabled: true
        }
    )
    {
        id
	title
	content
        author
	createdAt
        commentsEnabled
    }
}
```
Для просмотра поста по id:
```graphql
query findPost {
    post(
        id:"id-66"
    )
	{
        id
	title
	content
        author
	createdAt
        commentsEnabled
	}
}
```
Для просмотра постов (с пагинацией):
```graphql
query findPosts {
	posts(
	    limit:5,
	    offset:0
    )
    {
        id
	title
	content
        author
	createdAt
        commentsEnabled
	}
}
```
Для создания комментария к посту с id:
```graphql
mutation createComment {
    createComment(
        input: {
            postId:"id-66",
            content: "You have very cool post",
            author: "Gleb"
        }
    )
    {
        postId
        content
        author
        createdAt
    }
}
```
Для просмотра комментариев (с пагинацией) к посту с id:
```graphql
query findComments {
	comments(
	    postId: "id-66",
	    limit:0,
	    offset:0
    )
	{
        post{
            	id
    	  	title
    	  	content
        	author
    	  	createdAt
        	commentsEnabled
        }
        comments{
            	id
        	postId
        	parentId
        	content
        	author
        	createdAt
        }
	}
}
```
Для запрета добавления новых комментариев к посту с id:
```graphql
mutation makeCommentsDisable {
    disableComments(
        postId: "id-66"
    )
}
```
Для подписки на все новые комментарии к посту с id:
```graphql
subscription subscibeToPost{
    commentAdded(
        postId: "id-66"
    )
    {
        id
        postId
        parentId
        content
        author
        createdAt
    }
}
```

## Структура проекта


```bash
forum
    ├── internal
    │   ├── graphql
    │   │   ├── model
    │   │   │   └── models_gen.go
    │   │   ├── generated.go         # Автогенерируемый код gqlgen
    │   │   ├── resolver.go          # Реализация корневого резолвера
    │   │   ├── schema.graphql       # GraphQL схема
    │   │   └── schema.resolvers.go  # Реализация резолверов
    │   ├── storage
    │   │   ├── memory               # In-memory хранилище
    │   │   │   └── memory.go
    │   │   ├── postgres             # PostgreSQL хранилище
    │   │   │   └── postgres.go
    │   │   └── storage.go           # Интерфейсы хранилищ
    │   ├── gqlgen.yml
    │   └── tools.go
    ├── docker-compose.yml 
    ├── go.mod
    ├── go.sum
    └── main.go                      # Точка входа
```    
