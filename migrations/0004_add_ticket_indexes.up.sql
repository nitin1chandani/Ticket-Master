CREATE INDEX idx_tickets_event_status
ON tickets (event_id, status);

CREATE INDEX idx_tickets_booking_id
ON tickets (booking_id)
WHERE booking_id IS NOT NULL;

CREATE INDEX idx_tickets_reserved_expiry
ON tickets (reserved_until)
WHERE status = 'reserved';
