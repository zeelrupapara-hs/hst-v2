-- The group decides the currency an account is denominated in and how many digits it is shown to.
-- The account carried its own copy of the digits, which meant changing the group left every
-- account inside it on the old setting, and the currency already came from the group: one idea
-- with two sources that could disagree.
ALTER TABLE hst.accounts DROP COLUMN IF EXISTS currency_digits;

COMMENT ON TABLE hst.accounts IS 'money state per login; currency and its digits belong to the group';
