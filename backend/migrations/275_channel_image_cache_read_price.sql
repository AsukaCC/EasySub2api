ALTER TABLE channel_model_pricing
    ADD COLUMN IF NOT EXISTS image_cache_read_price NUMERIC(20,12);
