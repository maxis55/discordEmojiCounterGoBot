create table if not exists emojis
(
    emoji_id varchar(255) not null
        primary key,
    name     varchar(255),
    guild_id varchar(255),
    animated boolean
);

create index if not exists emojis_guild_id_emoji_id_index
    on emojis (guild_id, emoji_id);

create table if not exists messages
(
    message_id       varchar(255) not null
        primary key,
    content          text,
    guild_id         varchar(255),
    channel_id       varchar(255),
    author_id        varchar(255),
    flags            varchar(255),
    edited_timestamp timestamp,
    type             varchar(255),
    timestamp        timestamp
);

create table if not exists emoji_used
(
    id          bigserial
        primary key,
    message_id  varchar(255),
    m_author_id varchar(255),
    guild_id    varchar(255),
    channel_id  varchar(255),
    author_id   varchar(255),
    emoji_id    varchar(255),
    is_reaction boolean,
    timestamp   timestamp
);

create index if not exists emoji_used_m_author_id_index
    on emoji_used (m_author_id);

create index if not exists emoji_used_guild_id_index
    on emoji_used (guild_id);

create index if not exists emoji_used_author_id_index
    on emoji_used (author_id);

create index if not exists emoji_used_emoji_id_index
    on emoji_used (emoji_id);

create index if not exists emoji_used_channel_id_index
    on emoji_used (channel_id);

create index if not exists emoji_used_is_reaction_index
    on emoji_used (is_reaction);

create index if not exists emoji_used_timestamp_index
    on emoji_used (timestamp);

create table if not exists authors
(
    author_id   varchar(255) not null
        primary key,
    verified    boolean,
    username    varchar(255),
    global_name varchar(255),
    bot         boolean,
    system      boolean,
    mfa_enabled boolean
);

create table if not exists guilds
(
    guild_id          varchar(255) not null
        primary key,
    name              varchar(255),
    system_channel_id varchar(255),
    region            varchar(255),
    member_count      integer,
    icon              varchar(255),
    joined_at         timestamp,
    owner_id          varchar(255)
);

create table if not exists channels
(
    channel_id     varchar(255) not null
        primary key,
    owner_id       varchar(255),
    name           varchar(255),
    type           varchar(255),
    application_id varchar(255),
    parent_id      varchar(255),
    guild_id       varchar(255),
    nsfw           boolean,
    position       integer
);

create index if not exists channels_guild_id_index
    on channels (guild_id);

