--
-- PostgreSQL database cluster dump
--

SET default_transaction_read_only = off;

SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;

--
-- Roles
--

CREATE ROLE feeds;
ALTER ROLE feeds WITH SUPERUSER INHERIT CREATEROLE CREATEDB LOGIN REPLICATION BYPASSRLS PASSWORD 'SCRAM-SHA-256$4096:AFckvN8er6sw+Ct0ZbM1oA==$UyS/XK5O2e7e6jCKFlWkmHPyl/rjcFDNAmQsZYlL0LQ=:ABewV4wHdOEqu9Xod3hacoH0vt1UbUCy3TTpTiOJpOs=';

--
-- User Configurations
--








--
-- Databases
--

--
-- Database "template1" dump
--

\connect template1

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

--
-- PostgreSQL database dump complete
--

--
-- Database "feedstestdb" dump
--

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

--
-- Name: feedstestdb; Type: DATABASE; Schema: -; Owner: feeds
--

CREATE DATABASE feedstestdb WITH TEMPLATE = template0 ENCODING = 'UTF8' LOCALE_PROVIDER = libc LOCALE = 'en_US.utf8';


ALTER DATABASE feedstestdb OWNER TO feeds;

\connect feedstestdb

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
-- Name: feeds; Type: TABLE; Schema: public; Owner: feeds
--

CREATE TABLE public.feeds (
    id uuid NOT NULL,
    title text NOT NULL,
    url text,
    text text,
    type character(1),
    by character varying(16) NOT NULL,
    points smallint NOT NULL,
    n_comments smallint NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT feeds_type_check CHECK ((type = ANY (ARRAY['M'::bpchar, 'J'::bpchar, 'A'::bpchar, 'S'::bpchar])))
);


ALTER TABLE public.feeds OWNER TO feeds;

--
-- Name: hiddenfeeds; Type: TABLE; Schema: public; Owner: feeds
--

CREATE TABLE public.hiddenfeeds (
    id uuid NOT NULL,
    to_uname character varying(16) NOT NULL,
    feed uuid NOT NULL
);


ALTER TABLE public.hiddenfeeds OWNER TO feeds;

--
-- Data for Name: feeds; Type: TABLE DATA; Schema: public; Owner: feeds
--

COPY public.feeds (id, title, url, text, type, by, points, n_comments, created_at, deleted_at) FROM stdin;
cdff23df-b62c-446d-a2c1-b759fa342c5c	second first ask	localhost:8080/ask	is it first ask?	A	rootz	0	0	2025-10-26 14:41:46.724844+00	\N
e13cfca7-6bb0-4f67-ab96-e8941c20911a	hi from rootz	rootz.com/	anu rootz	S	rootz	0	0	2025-10-29 11:54:00.274446+00	\N
ab4f12b7-2143-4eda-b6ba-fd31840f9dbd	second feed	localhost:8080/second	this is second feed	M	root	0	0	2025-10-26 12:21:42.837155+00	2025-11-01 20:00:20.922913+00
b577c5cc-3dcd-42e8-ba21-387588dc6503	a	localhost:8083	take a look into my project	M	rootz	0	0	2025-11-01 20:04:26.142619+00	\N
42f9704d-2fba-437d-a3ab-b6da8ff8eb13	first feed	localhost:8080	this is first feed	M	root	0	0	2025-10-26 12:21:35.631518+00	2025-11-10 19:20:45.036242+00
14ed0f31-7a38-4c0b-ac48-8714378810f0	emm	anu.com	ra ruh	M	root	0	0	2025-11-10 20:01:55.764736+00	\N
\.


--
-- Data for Name: hiddenfeeds; Type: TABLE DATA; Schema: public; Owner: feeds
--

COPY public.hiddenfeeds (id, to_uname, feed) FROM stdin;
769d86bc-e0cf-4dd6-b89b-4aa932c64a2a	root	b577c5cc-3dcd-42e8-ba21-387588dc6503
\.


--
-- Name: feeds feeds_pkey; Type: CONSTRAINT; Schema: public; Owner: feeds
--

ALTER TABLE ONLY public.feeds
    ADD CONSTRAINT feeds_pkey PRIMARY KEY (id);


--
-- Name: hiddenfeeds hiddenfeeds_pkey; Type: CONSTRAINT; Schema: public; Owner: feeds
--

ALTER TABLE ONLY public.hiddenfeeds
    ADD CONSTRAINT hiddenfeeds_pkey PRIMARY KEY (id);


--
-- Name: hiddenfeeds unique_hidden_per_user; Type: CONSTRAINT; Schema: public; Owner: feeds
--

ALTER TABLE ONLY public.hiddenfeeds
    ADD CONSTRAINT unique_hidden_per_user UNIQUE (to_uname, feed);


--
-- Name: hiddenfeeds fk_feed; Type: FK CONSTRAINT; Schema: public; Owner: feeds
--

ALTER TABLE ONLY public.hiddenfeeds
    ADD CONSTRAINT fk_feed FOREIGN KEY (feed) REFERENCES public.feeds(id);


--
-- PostgreSQL database dump complete
--

--
-- Database "postgres" dump
--

\connect postgres

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

--
-- PostgreSQL database dump complete
--

--
-- PostgreSQL database cluster dump complete
--

