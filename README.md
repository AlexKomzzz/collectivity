## Веб сервис для для ведения учета оплаты и задолженности коллективного садоводства

Данный проект поднят на удаленном хосте при помощи docker.

Используется:
    Регистрация при помощи Oauth2 (Goolgle, Яндекс)
    Регистрация по эл. почте и паролю с подтверждением почты (отправка письма на почту клиента для подтверждения)
    Создание нового аккаунта
    Восстановление пароля с подтверждением по почте
    Учетная запись администратора и обычного клиента
    Для администратора возможность загружать exel таблицу, из которой берутся данные по клиентам
    Авторизация по JWT (переделано на сессии с cookie)
    БД - Postgres
    redis - для хранения временных данных
    Docker
    https://github.com/AlexKomzzz/collectivity-tlg-bot  бот для данного проекта

______________________________________




Ошибки:
Проверить работу, после PostForm и для бота. 
Переделать аутент на сессии с куками
    1. разделить HTML файлы и CSS, JS
    2. обработать ошибки при авторизации через другие соц сети


Сделать:
    - сделать сессии с куки
    - CI/CD на github action
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