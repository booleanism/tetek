--
-- PostgreSQL database dump
--

-- Dumped from database version 18.0
-- Dumped by pg_dump version 18.1 (Debian 18.1-1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: users; Type: TABLE; Schema: public; Owner: account
--

CREATE TABLE public.users (
    id uuid NOT NULL,
    uname character varying(16) NOT NULL,
    email character varying(64) NOT NULL,
    passwd text NOT NULL,
    role character(1) NOT NULL,
    created_at timestamp with time zone NOT NULL,
    CONSTRAINT users_role_check CHECK ((role = ANY (ARRAY['M'::bpchar, 'N'::bpchar])))
);


ALTER TABLE public.users OWNER TO account;

--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: account
--

COPY public.users (id, uname, email, passwd, role, created_at) FROM stdin;
dda10e28-4bc2-4add-91d8-5c4271565b2f	uname	uname@email	$2a$10$wuDFZw36p2Pd8Q7UPVoYbOL.cMmBZSyBLZAOVIf6K2eRAdAPJA2em	N	2025-10-23 21:53:26.319764+00
8e4bea94-2c69-423f-88bd-c42927d8cafa	root	root@localhost	$2a$10$O2quzPGc4mLiK5Ip8osszu8YULm4B/Qw0ILuGQjzz5u1d1hPW8CZy	N	2025-10-26 08:06:46.399555+00
4d12c328-346b-4e8e-ac1e-d3aa301db7f5	rootz	rootz@localhost	$2a$10$qbUPqjbuyZnxd4l30e4exePctJxsSyINGSYFi.l2bZEB/JBi24BPS	N	2025-10-26 14:10:51.406125+00
\.


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: account
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: account
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: users users_uname_key; Type: CONSTRAINT; Schema: public; Owner: account
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_uname_key UNIQUE (uname);


--
-- PostgreSQL database dump complete
--

