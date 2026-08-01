-- The engine counts in 1/100000000 of a lot now. Rows written before it carry only the coarse
-- number, so give each one its extended equivalent; the coarse column keeps its own meaning.
UPDATE hst.orders    SET volume_initial_ext = volume_initial * 10000 WHERE volume_initial_ext = 0;
UPDATE hst.orders    SET volume_current_ext = volume_current * 10000 WHERE volume_current_ext = 0;
UPDATE hst.positions SET volume_ext         = volume * 10000        WHERE volume_ext = 0;
UPDATE hst.deals     SET volume_ext         = volume * 10000        WHERE volume_ext = 0;
