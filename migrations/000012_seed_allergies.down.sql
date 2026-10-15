-- Soft-deletes rather than DELETE: user_allergies.allergy_id references these
-- rows, and the up migration skips names that already existed, so a hard
-- delete could remove catalog entries this migration never created.
UPDATE allergies
SET deleted_at = NOW()
WHERE deleted_at IS NULL
  AND (lower(name), category) IN (
    ('peanuts', 1), ('tree nuts', 1), ('shellfish', 1), ('fish', 1),
    ('milk', 1), ('eggs', 1), ('wheat', 1), ('soy', 1), ('sesame', 1),
    ('penicillin', 2), ('sulfa drugs', 2), ('aspirin', 2), ('nsaids', 2),
    ('local anesthetics', 2), ('contrast dye', 2),
    ('pollen', 3), ('dust mites', 3), ('mold', 3), ('pet dander', 3),
    ('insect stings', 3),
    ('latex', 4), ('nickel', 4), ('fragrances', 4), ('poison ivy', 4),
    ('other', 5)
);
