CREATE INDEX idx_posts_published_date ON posts (published, published_at DESC) 
WHERE published = true;