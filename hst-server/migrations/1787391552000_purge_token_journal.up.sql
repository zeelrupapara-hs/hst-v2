-- sign-in journal rows used to carry the issued tokens in detail; they are purged and no longer written
UPDATE hst.journal SET detail = NULL WHERE type = 12 AND code = 4;
