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

INSERT INTO tb_users (username, email, password_hash, date_registration, date_last_online)
VALUES
	('test1', 'test1@email.com', '\x92b2723f184a5f9b17ba52b88079391b', current_timestamp, current_timestamp);

-- test password: 1234567890qwerty
--SELECT decode ('92b2723f184a5f9b17ba52b88079391b', 'hex')