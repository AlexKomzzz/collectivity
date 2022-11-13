-- CREATE TABLE users
-- (
--     id              serial          not null unique  primary key,
--     id_google       varchar(255)    unique                      ,
--     id_yandex       varchar(255)    unique                      ,
--     username        varchar(255)    not null unique             , -- ФИО
--     first_name      varchar(255)    not null                    , -- имя
--     last_name       varchar(255)    not null                    , -- фамилия
--     middle_name     varchar(255)                                , -- отчество
--     password_hash   varchar(255)                                ,
--     email           varchar(255)    not null unique             ,
--     role_user       varchar(255)                                ,  -- роль (напр. admin)
--     debt            varchar(255)
-- );

CREATE TABLE users
(
    id              serial          not null unique  primary key,
    username        varchar(255)    not null unique             , -- ФИО
    first_name      varchar(255)    not null                    , -- имя
    last_name       varchar(255)    not null                    , -- фамилия
    middle_name     varchar(255)                                , -- отчество
    role_user       varchar(255)                                ,  -- роль (напр. admin)
    debt            varchar(255)
);

CREATE TABLE auth
(
    id_user         int references users (id) on delete cascade         not null unique,
    id_google       varchar(255)    unique                      ,
    id_yandex       varchar(255)    unique                      ,
    password_hash   varchar(255)    not null                    ,
    email           varchar(255)    not null unique             
);

CREATE TABLE authdata
(
    id              serial          not null unique  primary key,
    first_name      varchar(255)    not null                    , -- имя
    last_name       varchar(255)    not null                    , -- фамилия
    middle_name     varchar(255)                                , -- отчество
    password_hash   varchar(255)    not null                    ,
    email           varchar(255)    not null unique             
);

-- CREATE TABLE debts
-- (
--     id          serial                                              not null unique  primary key,
--     id_user     int references users (id) on delete cascade         not null unique,
--     debt        varchar(255)
-- );

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