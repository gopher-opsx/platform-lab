CREATE TABLE IF NOT EXISTS notifications (
    id TEXT PRIMARY KEY,
    order_id TEXT NOT NULL,
    source_event_id TEXT NOT NULL UNIQUE,
    source_event_type TEXT NOT NULL,
    channel TEXT NOT NULL,
    recipient TEXT NOT NULL,
    subject TEXT NOT NULL,
    body TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('delivered', 'failed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    delivered_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_notifications_order_id ON notifications(order_id);

CREATE TABLE IF NOT EXISTS processed_events (
    event_id TEXT PRIMARY KEY,
    event_type TEXT NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_notification_processed_at ON processed_events(processed_at);

CREATE TABLE IF NOT EXISTS delivery_log (
    id BIGSERIAL PRIMARY KEY,
    notification_id TEXT NOT NULL REFERENCES notifications(id),
    attempt_number INTEGER NOT NULL CHECK (attempt_number > 0),
    status TEXT NOT NULL CHECK (status IN ('delivered', 'failed')),
    provider TEXT NOT NULL,
    message TEXT,
    attempted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (notification_id, attempt_number)
);
CREATE INDEX IF NOT EXISTS idx_delivery_log_notification ON delivery_log(notification_id);
