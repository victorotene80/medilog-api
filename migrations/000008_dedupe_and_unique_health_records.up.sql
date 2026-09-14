-- Enforce, in the database, the rules the application already applies:
--   * one allergy per (user, case-insensitive name)
--   * one emergency contact per (user, phone)
--   * at most one primary emergency contact per user
--
-- Until now these were application-only and therefore racy: two concurrent
-- requests could both pass the "does it already exist?" check and both insert.
--
-- Each backfill MUST run before its index, or CREATE UNIQUE INDEX aborts on the
-- first existing duplicate. Duplicates are soft-deleted rather than removed,
-- keeping with the soft-delete convention used throughout the schema; the
-- oldest row (lowest id) always wins.

-- 1. user_allergies: collapse case-insensitive duplicate names per user.
UPDATE user_allergies ua
   SET deleted_at = NOW()
 WHERE ua.deleted_at IS NULL
   AND EXISTS (
       SELECT 1
         FROM user_allergies keep
        WHERE keep.deleted_at IS NULL
          AND keep.user_id = ua.user_id
          AND lower(keep.name) = lower(ua.name)
          AND keep.id < ua.id);

-- lower(name) matches how UserAllergyRepository.FindByUserIDAndName compares,
-- so the query and the constraint agree and the index is actually used.
CREATE UNIQUE INDEX IF NOT EXISTS ux_user_allergies_user_name
    ON user_allergies (user_id, lower(name))
 WHERE deleted_at IS NULL;

-- 2. emergency_contacts: collapse duplicate phone numbers per user.
UPDATE emergency_contacts ec
   SET deleted_at = NOW()
 WHERE ec.deleted_at IS NULL
   AND EXISTS (
       SELECT 1
         FROM emergency_contacts keep
        WHERE keep.deleted_at IS NULL
          AND keep.user_id = ec.user_id
          AND keep.phone = ec.phone
          AND keep.id < ec.id);

CREATE UNIQUE INDEX IF NOT EXISTS ux_emergency_contacts_user_phone
    ON emergency_contacts (user_id, phone)
 WHERE deleted_at IS NULL;

-- 3a. Demote extra primaries, keeping the oldest.
UPDATE emergency_contacts ec
   SET is_primary = false, updated_at = NOW()
 WHERE ec.deleted_at IS NULL
   AND ec.is_primary
   AND EXISTS (
       SELECT 1
         FROM emergency_contacts keep
        WHERE keep.deleted_at IS NULL
          AND keep.is_primary
          AND keep.user_id = ec.user_id
          AND keep.id < ec.id);

-- 3b. Promote the oldest contact for users left with no primary at all.
--
-- This is not cosmetic. The REST layer used to drop is_primary when building
-- the create command, so every contact was stored with is_primary = false and
-- most existing users have zero primaries. Without this step GetUserHandler
-- keeps falling through to its "first by id" fallback forever.
UPDATE emergency_contacts ec
   SET is_primary = true, updated_at = NOW()
 WHERE ec.deleted_at IS NULL
   AND ec.id = (
       SELECT MIN(inner_ec.id)
         FROM emergency_contacts inner_ec
        WHERE inner_ec.deleted_at IS NULL
          AND inner_ec.user_id = ec.user_id)
   AND NOT EXISTS (
       SELECT 1
         FROM emergency_contacts p
        WHERE p.deleted_at IS NULL
          AND p.is_primary
          AND p.user_id = ec.user_id);

CREATE UNIQUE INDEX IF NOT EXISTS ux_emergency_contacts_one_primary
    ON emergency_contacts (user_id)
 WHERE is_primary AND deleted_at IS NULL;
