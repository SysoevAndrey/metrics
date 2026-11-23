CREATE TABLE metrics (
  id TEXT NOT NULL,
  mtype TEXT NOT NULL,
  delta BIGINT,
  value DOUBLE PRECISION,
  PRIMARY KEY (id, mtype),
  CHECK (mtype IN ('counter', 'gauge'))
);