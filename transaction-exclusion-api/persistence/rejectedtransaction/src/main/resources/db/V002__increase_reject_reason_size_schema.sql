-- =======================================================
-- rejected_transactions table
-- =======================================================
alter table if exists rejected_transactions
  alter column reject_reason type varchar(512);
