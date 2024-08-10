create table if not exists emojis
(
    emoji_id varchar(255) not null
        primary key,
    name     varchar(255) null,
    guild_id varchar(255) null,
    animated bool         null
);

create table if not exists messages
(
    message_id       varchar(255) not null
        primary key,
    content          text         null,
    guild_id         varchar(255) null,
    channel_id       varchar(255) null,
    author_id        varchar(255) null,
    flags            varchar(255) null,
    edited_timestamp timestamp    null,
    type             varchar(255) null,
    timestamp        timestamp    null
);


create table if not exists emoji_used
(
    id          bigserial primary key
        message_id varchar (255) null,
    guild_id    varchar(255) null,
    channel_id  varchar(255) null,
    author_id   varchar(255) null,
    emoji_id    varchar(255) null,
    is_reaction bool         null,
);

create table if not exists authors
(
    author_id   varchar(255) not null
        primary key,
    verified    bool         null,
    username    varchar(255) null,
    global_name varchar(255) null,
    bot         bool         null,
    system      bool         null,
    mfa_enabled bool         null
);

create table if not exists guilds
(
    guild_id          varchar(255) not null
        primary key,
    name              varchar(255) null,
    system_channel_id varchar(255) null,
    region            varchar(255) null,
    member_count      int          null,
    icon              varchar(255) null,
    joined_at         timestamp    null,
    owner_id          varchar(255) null
);


create table if not exists channels
(
    channel_id     varchar(255) not null
        primary key,
    owner_id       varchar(255) null,
    name           varchar(255) null,
    type           varchar(255) null,
    application_id varchar(255) null,
    parent_id      varchar(255) null,
    guild_id       varchar(255) null,
    nsfw           bool         null,
    position       int          null
);