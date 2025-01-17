-- =======================================================
-- full_transactions table
-- =======================================================
create table if not exists full_transactions (
  tx_hash                       bytea,
  tx_rlp                        bytea,
  primary key (tx_hash)
);

-- =======================================================
-- rejected_transactions table
-- =======================================================
create table if not exists rejected_transactions (
  created_epoch_milli          bigint,
  tx_hash                      bytea,
  tx_from                      bytea,
  tx_to                        bytea,
  tx_nonce                     bigint,
  reject_stage                 varchar(5), -- SEQ,RPC,P2P
  reject_reason                varchar(256),
  reject_timestamp             bigint,
  block_number                 bigint,
  overflows                    jsonb,
  primary key (tx_hash, reject_reason),
  foreign key (tx_hash) references full_transactions
);
