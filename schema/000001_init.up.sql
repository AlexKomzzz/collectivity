CREATE TABLE users
(
    id              serial          not null unique  primary key,
    id_google       varchar(255)    unique                      ,
    id_yandex       varchar(255)    unique                      ,
    first_name      varchar(255)    not null                    , -- имя
    last_name       varchar(255)    not null                    , -- фамилия
    patronymic      varchar(255)                                , -- отчество
    password_hash   varchar(255)                                ,
    email           varchar(255)    not null unique             ,
    role            varchar(255)                                  -- роль (напр. Админ)
);

-- CREATE TABLE todo_lists
-- (
--     id              serial          not null unique,
--     title           varchar(255)    not null,
--     description     varchar(255)
-- );

-- CREATE TABLE user_lists
-- (
--     id          serial                                              not null unique,
--     user_id     int references users (id) on delete cascade         not null,
--     list_id     int references todo_lists (id) on delete cascade    not null
-- );

-- CREATE TABLE todo_items
-- (
--     id              serial          not null unique,
--     title           varchar(255)    not null,
--     description     varchar(255),
--     done            boolean         not null default false
-- );

-- CREATE TABLE lists_items
-- (
--     id          serial                                              not null unique,
--     item_id     int references todo_items (id) on delete cascade         not null,
--     list_id     int references todo_lists (id) on delete cascade    not null
-- );