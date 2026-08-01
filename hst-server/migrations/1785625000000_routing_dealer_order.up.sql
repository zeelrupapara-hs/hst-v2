-- the Dealers tab is an ordered list: the request goes to the first dealer who can take it
ALTER TABLE hst.routing_dealers ADD COLUMN IF NOT EXISTS dealer_index INTEGER NOT NULL DEFAULT 0;

UPDATE hst.routing_dealers d
   SET dealer_index = o.rn - 1
  FROM (SELECT dealer_id, row_number() OVER (PARTITION BY routing_id ORDER BY dealer_id) AS rn
          FROM hst.routing_dealers) o
 WHERE o.dealer_id = d.dealer_id;

CREATE UNIQUE INDEX IF NOT EXISTS routing_dealers_order_uidx
    ON hst.routing_dealers (routing_id, dealer_index);
