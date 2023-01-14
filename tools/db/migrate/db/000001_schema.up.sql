CREATE TABLE public.tb_users (
    id serial NOT NULL,
    username varchar(255) NOT NULL,
    email varchar(255) NOT NULL,
    password_hash bytea NOT NULL,
    date_registration timestamptz NOT NULL,
    date_last_online timestamptz NOT NULL,
    CONSTRAINT tb_users_pk PRIMARY KEY (id),
    CONSTRAINT tb_users_username UNIQUE (username),
    CONSTRAINT tb_users_email UNIQUE (email)
);