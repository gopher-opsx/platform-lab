INSERT INTO products (
    id,
    name,
    description,
    price_cents,
    currency,
    image_url,
    in_stock
)
VALUES
    (
        'prod-001',
        'CloudBook Pro 14',
        '14-inch developer laptop with 32 GB RAM and 1 TB SSD.',
        169900,
        'USD',
        '/images/cloudbook-pro-14.jpg',
        TRUE
    ),
    (
        'prod-002',
        'CloudPhone X',
        '5G smartphone with 256 GB storage and OLED display.',
        89900,
        'USD',
        '/images/cloudphone-x.jpg',
        TRUE
    ),
    (
        'prod-003',
        'CloudPods',
        'Wireless noise-cancelling earbuds with charging case.',
        19900,
        'USD',
        '/images/cloudpods.jpg',
        FALSE
    )
ON CONFLICT (id) DO UPDATE
SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    price_cents = EXCLUDED.price_cents,
    currency = EXCLUDED.currency,
    image_url = EXCLUDED.image_url,
    in_stock = EXCLUDED.in_stock,
    updated_at = NOW();
