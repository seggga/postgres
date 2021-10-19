BEGIN;

DROP INDEX videos_created_at_idx;
DROP INDEX comments_created_at_idx;
DROP INDEX videos_caption_idx;

COMMIT;