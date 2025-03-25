CREATE TABLE IF NOT EXISTS posts (
    id TEXT PRIMARY KEY,
    author TEXT NOT NULL,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    comments_enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS comments (
    id TEXT PRIMARY KEY,
    post_id TEXT NOT NULL,
    parent_id TEXT NOT NULL,
    author TEXT NOT NULL,
    content TEXT CHECK (LENGTH(content) <= 2000),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- CREATE INDEX ON comments (post_id, parent_id);