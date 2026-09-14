-- Bound OTP guesses per code.
--
-- Verification previously mutated nothing on a wrong guess, so a live code
-- accepted unlimited attempts and the only ceiling was the IP-keyed route
-- limiter. That bounds guesses per caller, not per code, so a pool of addresses
-- could work through the 6-digit keyspace of a password-reset code well inside
-- its TTL. The counter has to live with the code for the limit to mean anything.
ALTER TABLE otp_codes
    ADD COLUMN IF NOT EXISTS attempts INTEGER NOT NULL DEFAULT 0;

-- No index is added here. The verification lookup filters on
-- (recipient, purpose, used_at IS NULL) ordered by created_at DESC, which is
-- already served by idx_otp_codes_recipient_purpose_created from migration
-- 000005. A second index on the same leading columns would only add write cost
-- on a table that takes an insert per OTP request.
