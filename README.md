## Веб сервис для для ведения учета оплаты и задолженности коллективного садоводства

Используется:
    Регистрация при помощи Oauth2 (Goolgle, Яндекс)
    Регистрация по эл. почте и паролю с подтверждением почты
    Восстановление пароля
    Учетная запись администратора
    JWT
    Postgres
    redis
    Docker

1. настроить авторизацию клиента по Gmail (Яндекс почте)
2. иметь учетку с правами админа для загрузки таблицы с задолженностями
3. парсить загруженную таблицу в БД
4. создать пользователя в БД. пароль для БД передавать в Environment
5. Создать телеграмм бота (доп.)
6. Нужен ли Nginx? для https
7. настроить VDS, получить домен
8. развернуть докером
9. для идентификации почты генерировать токен с другой солью


Ошибки:
Проверить работу, после PostForm и для бота. 
Переделать аутент на сессии с куками
    1. разделить HTML файлы и CSS, JS
    2. обработать ошибки при авторизации через другие соц сети


Сделать:
    - написать скрипт для клиента, при получении JWT формировать запрос на стратовую страницу с заголовком авторизации
    - скрипт, который создает клиента с ролью Админа в БД
    

### Миграции в БД

создание файлов миграции

    $ migrate create -ext sql -dir ./schema -seq init

применить файл UP

    $ migrate -path ./schema -database 'postgres://postgres:qwerty@localhost:5432/postgres?sslmode=disable' up (down)

Создать пользователя админа:

    $ INSERT INTO users (first_name, last_name, email, role) VALUES (admin, admin, admin@admin.com, admin)


### docker

развернуть контейнер с приложением

    $ docker compose up -d --build
    $ docker compose --env-file .env up --build

зайти в оболочку bash контейнера postgres

    $ docker exec -it db /bin/bash
    $ psql -U [userDB]


### создать админа 

в среде окружения передать пароль для админа
перейти на http://domen/auth/admin для создания админа


SSL сертификаты для localhost созданы при помощи:

    $ https://github.com/FiloSottile/mkcert