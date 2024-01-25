-- Create database
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT FROM pg_database WHERE datname = 'todobukh') THEN
        CREATE DATABASE todobukh;
    END IF;
END $$;

\c todobukh;

-- Create schema
CREATE SCHEMA IF NOT EXISTS public;

-- Create users table
CREATE TABLE IF NOT EXISTS public."user" (
    user_id serial PRIMARY KEY,
    username character varying(255) NOT NULL,
    CONSTRAINT user_un UNIQUE (username)
);

-- Create credentials table
CREATE TABLE IF NOT EXISTS public.credentials (
    user_id integer NOT NULL,
    password character varying(60) NOT NULL,
    PRIMARY KEY (user_id),
    CONSTRAINT credentials_fk FOREIGN KEY (user_id) REFERENCES public."user" (user_id)
);

-- Create task table
CREATE TABLE IF NOT EXISTS public.task (
    user_id integer NOT NULL,
    task_description character varying(2000) NOT NULL,
    task_id serial PRIMARY KEY,
    is_completed boolean DEFAULT false NOT NULL,
    CONSTRAINT task_fk FOREIGN KEY (user_id) REFERENCES public."user" (user_id)
);

-- Set sequence values
-- SELECT pg_catalog.setval('public."user_user_id_seq"', 1, false);
-- SELECT pg_catalog.setval('public.credentials_user_id_seq', 1, false);
-- SELECT pg_catalog.setval('public.task_task_id_seq', 1, false);

-- Create primary key constraints
-- ALTER TABLE ONLY public."user" ADD CONSTRAINT user_pk PRIMARY KEY (user_id);
-- ALTER TABLE ONLY public.credentials ADD CONSTRAINT credentials_pk PRIMARY KEY (user_id);
-- ALTER TABLE ONLY public.task ADD CONSTRAINT task_pk PRIMARY KEY (task_id);

-- Create foreign key constraints
-- ALTER TABLE ONLY public.credentials ADD CONSTRAINT credentials_fk FOREIGN KEY (user_id) REFERENCES public."user" (user_id);
-- ALTER TABLE ONLY public.task ADD CONSTRAINT task_fk FOREIGN KEY (user_id) REFERENCES public."user" (user_id);

