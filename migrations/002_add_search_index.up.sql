-- Add GIN index for full-text search on message content
CREATE INDEX idx_messages_content_search ON messages USING GIN (to_tsvector('english', content));
