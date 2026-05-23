-- Discord snowflake IDs today are at most 19 chars (decimal int64). varchar(32)
-- gives generous headroom in case Discord ever changes the ID format, while
-- still documenting "this is a short ID, not arbitrary text" vs varchar(255).
-- Postgres only scans the table to validate no row exceeds the new length; for
-- real snowflakes this is fast and the table is not rewritten.

ALTER TABLE emojis     ALTER COLUMN emoji_id    TYPE varchar(32);
ALTER TABLE emojis     ALTER COLUMN guild_id    TYPE varchar(32);

ALTER TABLE messages   ALTER COLUMN message_id  TYPE varchar(32);
ALTER TABLE messages   ALTER COLUMN guild_id    TYPE varchar(32);
ALTER TABLE messages   ALTER COLUMN channel_id  TYPE varchar(32);
ALTER TABLE messages   ALTER COLUMN author_id   TYPE varchar(32);

ALTER TABLE emoji_used ALTER COLUMN message_id  TYPE varchar(32);
ALTER TABLE emoji_used ALTER COLUMN m_author_id TYPE varchar(32);
ALTER TABLE emoji_used ALTER COLUMN guild_id    TYPE varchar(32);
ALTER TABLE emoji_used ALTER COLUMN channel_id  TYPE varchar(32);
ALTER TABLE emoji_used ALTER COLUMN author_id   TYPE varchar(32);
ALTER TABLE emoji_used ALTER COLUMN emoji_id    TYPE varchar(32);

ALTER TABLE authors    ALTER COLUMN author_id   TYPE varchar(32);

ALTER TABLE guilds     ALTER COLUMN guild_id          TYPE varchar(32);
ALTER TABLE guilds     ALTER COLUMN system_channel_id TYPE varchar(32);
ALTER TABLE guilds     ALTER COLUMN owner_id          TYPE varchar(32);

ALTER TABLE channels   ALTER COLUMN channel_id     TYPE varchar(32);
ALTER TABLE channels   ALTER COLUMN owner_id       TYPE varchar(32);
ALTER TABLE channels   ALTER COLUMN application_id TYPE varchar(32);
ALTER TABLE channels   ALTER COLUMN parent_id      TYPE varchar(32);
ALTER TABLE channels   ALTER COLUMN guild_id       TYPE varchar(32);
