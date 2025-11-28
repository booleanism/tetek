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
-- Name: comments; Type: TABLE; Schema: public; Owner: comms
--

CREATE TABLE public.comments (
    id uuid NOT NULL,
    parent uuid NOT NULL,
    text text NOT NULL,
    by character varying(16) NOT NULL,
    created_at timestamp with time zone NOT NULL
);


ALTER TABLE public.comments OWNER TO comms;

--
-- Name: points; Type: TABLE; Schema: public; Owner: comms
--

CREATE TABLE public.points (
    id uuid NOT NULL,
    by character varying(16) NOT NULL,
    comments_id uuid NOT NULL
);


ALTER TABLE public.points OWNER TO comms;

--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: comms
--

CREATE TABLE public.schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL
);


ALTER TABLE public.schema_migrations OWNER TO comms;

--
-- Data for Name: comments; Type: TABLE DATA; Schema: public; Owner: comms
--

COPY public.comments (id, parent, text, by, created_at) FROM stdin;
07099af9-3c49-4f55-85e7-54dfc11ea138	14ed0f31-7a38-4c0b-ac48-8714378810f0	haloo!	root	2025-11-16 13:41:58.177373+00
7d45778e-0fd7-4ff2-8ca9-082d80ec7f7c	07099af9-3c49-4f55-85e7-54dfc11ea138	haloo there!	root	2025-11-16 14:04:12.272858+00
872a602a-52ea-4525-a5f8-c8ad0958708b	14ed0f31-7a38-4c0b-ac48-8714378810f0	hello from rootz	rootz	2025-11-24 15:38:31.769862+00
\.


--
-- Data for Name: points; Type: TABLE DATA; Schema: public; Owner: comms
--

COPY public.points (id, by, comments_id) FROM stdin;
\.


--
-- Data for Name: schema_migrations; Type: TABLE DATA; Schema: public; Owner: comms
--

COPY public.schema_migrations (version, dirty) FROM stdin;
2	f
\.


--
-- Name: comments comments_pkey; Type: CONSTRAINT; Schema: public; Owner: comms
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_pkey PRIMARY KEY (id);


--
-- Name: points point_per_user; Type: CONSTRAINT; Schema: public; Owner: comms
--

ALTER TABLE ONLY public.points
    ADD CONSTRAINT point_per_user UNIQUE (by, comments_id);


--
-- Name: points points_pkey; Type: CONSTRAINT; Schema: public; Owner: comms
--

ALTER TABLE ONLY public.points
    ADD CONSTRAINT points_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: comms
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- PostgreSQL database dump complete
--
