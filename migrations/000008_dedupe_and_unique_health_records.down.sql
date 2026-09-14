DROP INDEX IF EXISTS ux_emergency_contacts_one_primary;
DROP INDEX IF EXISTS ux_emergency_contacts_user_phone;
DROP INDEX IF EXISTS ux_user_allergies_user_name;

-- The soft-deletes and primary-flag promotions from the up migration are
-- intentionally NOT reverted: the de-duplicated data is the correct state, and
-- resurrecting duplicates would only re-break the indexes on the next up.
