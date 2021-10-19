BEGIN;

CREATE INDEX concurrently videos_created_at_idx ON videos USING BTREE (created_at);
CREATE INDEX concurrently comments_created_at_idx ON comments USING BTREE (created_at);
CREATE INDEX concurrently videos_caption_idx ON videos USING BTREE (caption text_pattern_ops);

COMMIT;