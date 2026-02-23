-- Up
ALTER TABLE outbox_events
ADD COLUMN request_id VARCHAR(255);

-- Tambahkan index agar pencarian log berdasarkan request_id di DB menjadi cepat
CREATE INDEX idx_outbox_request_id ON outbox_events (request_id);