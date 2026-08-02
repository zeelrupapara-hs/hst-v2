-- The routes now demand every right a permission depends on, not just the nearest one. Manager
-- rows written before that ran under the looser rule, so any holding a dependent right without
-- its prerequisite would lose access on deploy without this.

-- reading a trade needs account access, so a manager already reading trades keeps doing so
UPDATE hst.managers
   SET right_acc_read = 1
 WHERE right_acc_read = 0
   AND (right_trades_read = 1 OR right_trades_manager = 1 OR right_trades_dealer = 1);

-- dealing and writing a trade both sit on top of reading one
UPDATE hst.managers
   SET right_trades_read = 1
 WHERE right_trades_read = 0
   AND (right_trades_manager = 1 OR right_trades_dealer = 1);

-- deleting an account needs the right that edits one
UPDATE hst.managers
   SET right_acc_manager = 1
 WHERE right_acc_manager = 0
   AND right_acc_delete = 1;
