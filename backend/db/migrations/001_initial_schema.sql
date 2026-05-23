CREATE TABLE IF NOT EXISTS emojis (
    emoji_id varchar(255) NOT NULL PRIMARY KEY,
    name     varchar(255),
    guild_id varchar(255),
    animated boolean
);

CREATE INDEX IF NOT EXISTS emojis_guild_id_emoji_id_index ON emojis (guild_id, emoji_id);

CREATE TABLE IF NOT EXISTS messages (
    message_id       varchar(255) NOT NULL PRIMARY KEY,
    content          text,
    guild_id         varchar(255),
    channel_id       varchar(255),
    author_id        varchar(255),
    flags            varchar(255),
    edited_timestamp timestamp,
    type             varchar(255),
    timestamp        timestamp
);

CREATE TABLE IF NOT EXISTS emoji_used (
    id          bigserial PRIMARY KEY,
    message_id  varchar(255),
    m_author_id varchar(255),
    guild_id    varchar(255),
    channel_id  varchar(255),
    author_id   varchar(255),
    emoji_id    varchar(255),
    is_reaction boolean,
    timestamp   timestamp
);

CREATE INDEX IF NOT EXISTS emoji_used_m_author_id_index ON emoji_used (m_author_id);
CREATE INDEX IF NOT EXISTS emoji_used_guild_id_index    ON emoji_used (guild_id);
CREATE INDEX IF NOT EXISTS emoji_used_author_id_index   ON emoji_used (author_id);
CREATE INDEX IF NOT EXISTS emoji_used_emoji_id_index    ON emoji_used (emoji_id);
CREATE INDEX IF NOT EXISTS emoji_used_channel_id_index  ON emoji_used (channel_id);
CREATE INDEX IF NOT EXISTS emoji_used_is_reaction_index ON emoji_used (is_reaction);
CREATE INDEX IF NOT EXISTS emoji_used_timestamp_index   ON emoji_used (timestamp);
CREATE INDEX IF NOT EXISTS emoji_used_message_id_index  ON emoji_used (message_id);

CREATE TABLE IF NOT EXISTS authors (
    author_id   varchar(255) NOT NULL PRIMARY KEY,
    verified    boolean,
    username    varchar(255),
    global_name varchar(255),
    bot         boolean,
    system      boolean,
    mfa_enabled boolean
);

CREATE TABLE IF NOT EXISTS guilds (
    guild_id          varchar(255) NOT NULL PRIMARY KEY,
    name              varchar(255),
    system_channel_id varchar(255),
    region            varchar(255),
    member_count      integer,
    icon              varchar(255),
    joined_at         timestamp,
    owner_id          varchar(255)
);

CREATE TABLE IF NOT EXISTS channels (
    channel_id     varchar(255) NOT NULL PRIMARY KEY,
    owner_id       varchar(255),
    name           varchar(255),
    type           varchar(255),
    application_id varchar(255),
    parent_id      varchar(255),
    guild_id       varchar(255),
    nsfw           boolean,
    position       integer
);

CREATE INDEX IF NOT EXISTS channels_guild_id_index ON channels (guild_id);
