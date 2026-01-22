CREATE TABLE segment_stats (
   id BIGSERIAL PRIMARY KEY,

   stat_date DATE NOT NULL,
   ticker TEXT NOT NULL,
   pattern_type TEXT NOT NULL,

   segment_from TIMESTAMP NOT NULL,
   segment_to TIMESTAMP NOT NULL,

   total_matches INT NOT NULL,
   price_change NUMERIC,
   probability NUMERIC,

   created_at TIMESTAMP NOT NULL DEFAULT now(),
   updated_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX uniq_segment_stats_day
    ON segment_stats (stat_date, ticker, pattern_type);

