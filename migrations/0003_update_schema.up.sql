ALTER TABLE tickets
ADD COLUMN reserved_until TIMESTAMPTZ;

ALTER TABLE bookings
ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'pending';

ALTER TABLE bookings
ADD CONSTRAINT booking_status_check
CHECK (status IN ('pending', 'confirmed', 'expired', 'cancelled'));
