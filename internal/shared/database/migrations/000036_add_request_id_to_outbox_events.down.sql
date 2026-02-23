-- Down
DROP INDEX IF EXISTS idx_outbox_request_id;
ALTER TABLE outbox_events 
DROP COLUMN request_id;