-- 000007: coupons + coupon_uses tables

CREATE TABLE IF NOT EXISTS coupons (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code             VARCHAR(100) UNIQUE NOT NULL,
    discount_percent NUMERIC(5, 2) NOT NULL CHECK (discount_percent > 0 AND discount_percent <= 100),
    max_uses         INT NOT NULL DEFAULT 0,  -- 0 = unlimited
    used_count       INT NOT NULL DEFAULT 0,
    expires_at       TIMESTAMP WITH TIME ZONE NOT NULL,
    is_active        BOOLEAN NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_coupons_code ON coupons(code);

CREATE TABLE IF NOT EXISTS coupon_uses (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    coupon_id  UUID NOT NULL REFERENCES coupons(id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users(id)   ON DELETE CASCADE,
    order_id   UUID NOT NULL REFERENCES orders(id)  ON DELETE CASCADE,
    used_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(coupon_id, user_id)  -- one use per user per coupon
);

CREATE INDEX IF NOT EXISTS idx_coupon_uses_coupon ON coupon_uses(coupon_id);
CREATE INDEX IF NOT EXISTS idx_coupon_uses_user   ON coupon_uses(user_id);
