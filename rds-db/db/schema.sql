SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: identity; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA identity;

--
-- Name: workflow; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA workflow;

--
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


SET default_tablespace = '';

SET default_table_access_method = heap;


--
-- Name: api_client; Type: TABLE; Schema: identity; Owner: -
--

CREATE TABLE identity.api_client (
    id character varying(36) NOT NULL,
    organization_id character varying(36) NOT NULL,
    provider character varying(30) NOT NULL,
    info jsonb NOT NULL,
    role character varying(30) NOT NULL,
    secret text DEFAULT ''::text NOT NULL,
    creation_time timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    last_update_time timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    type character varying(30) DEFAULT ''::character varying NOT NULL,
    api_key_hash character varying(100) DEFAULT ''::character varying NOT NULL
);
--
-- Name: integration; Type: TABLE; Schema: identity; Owner: -
--

CREATE TABLE identity.integration (
    organization_id character varying(36) NOT NULL,
    partner character varying(50) NOT NULL,
    service character varying(50) NOT NULL,
    status character varying(30) NOT NULL,
    info jsonb NOT NULL,
    creation_time timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_by character varying(36) NOT NULL,
    last_update_time timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    last_updated_by character varying(36) NOT NULL
);


--
-- Name: organization; Type: TABLE; Schema: identity; Owner: -
--

CREATE TABLE identity.organization (
    id character varying(36) NOT NULL,
    name character varying(300) NOT NULL,
    email_domain character varying(100) NOT NULL,
    website character varying(300) NOT NULL,
    description text NOT NULL,
    creation_time timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    status character varying(30) NOT NULL,
    last_update_time timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    allowed_auth_methods character varying(30)[],
    created_by character varying(36) NOT NULL,
    auth_id character varying(36) DEFAULT ''::character varying NOT NULL,
    info jsonb DEFAULT '{}'::jsonb NOT NULL
);


--
-- Name: role; Type: TABLE; Schema: identity; Owner: -
--

CREATE TABLE identity.role (
    organization_id character varying(36) NOT NULL,
    id character varying(36) NOT NULL,
    name character varying(50) NOT NULL,
    description character varying(300) DEFAULT ''::character varying NOT NULL,
    permissions character varying(50)[] DEFAULT '{}'::character varying[] NOT NULL,
    creation_time timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


--
-- Name: user; Type: TABLE; Schema: identity; Owner: -
--

CREATE TABLE identity."user" (
    id character varying(36) NOT NULL,
    first_name character varying(30) NOT NULL,
    last_name character varying(30) NOT NULL,
    email character varying(100) NOT NULL,
    creation_time timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    last_update_time timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


--
-- Name: user_organization; Type: TABLE; Schema: identity; Owner: -
--

CREATE TABLE identity.user_organization (
    user_id character varying(36) NOT NULL,
    organization_id character varying(36) NOT NULL,
    user_role character varying(36) NOT NULL,
    allowed_auth_methods character varying(30)[],
    creation_time timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    last_update_time timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);



--
-- Name: auth_identity; Type: TABLE; Schema: workflow; Owner: -
--

CREATE TABLE workflow.auth_identity (
    "userId" uuid,
    "providerId" character varying(64) NOT NULL,
    "providerType" character varying(32) NOT NULL,
    "createdAt" timestamp(3) with time zone DEFAULT CURRENT_TIMESTAMP(3) NOT NULL,
    "updatedAt" timestamp(3) with time zone DEFAULT CURRENT_TIMESTAMP(3) NOT NULL
);


--
-- Name: auth_provider_sync_history; Type: TABLE; Schema: workflow; Owner: -
--

CREATE TABLE workflow.auth_provider_sync_history (
    id integer NOT NULL,
    "providerType" character varying(32) NOT NULL,
    "runMode" text NOT NULL,
    status text NOT NULL,
    "startedAt" timestamp(3) with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "endedAt" timestamp(3) with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    scanned integer NOT NULL,
    created integer NOT NULL,
    updated integer NOT NULL,
    disabled integer NOT NULL,
    error text
);


--
-- Name: auth_provider_sync_history_id_seq; Type: SEQUENCE; Schema: workflow; Owner: -
--

CREATE SEQUENCE workflow.auth_provider_sync_history_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: auth_provider_sync_history_id_seq; Type: SEQUENCE OWNED BY; Schema: workflow; Owner: -
--

ALTER SEQUENCE workflow.auth_provider_sync_history_id_seq OWNED BY workflow.auth_provider_sync_history.id;


--
-- Name: credentials_entity; Type: TABLE; Schema: workflow; Owner: -
--

CREATE TABLE workflow.credentials_entity (
    name character varying(128) NOT NULL,
    data text NOT NULL,
    type character varying(128) NOT NULL,
    "nodesAccess" json NOT NULL,
    "createdAt" timestamp(3) with time zone DEFAULT CURRENT_TIMESTAMP(3) NOT NULL,
    "updatedAt" timestamp(3) with time zone DEFAULT CURRENT_TIMESTAMP(3) NOT NULL,
    id character varying(36) NOT NULL
);


--
-- Name: event_destinations; Type: TABLE; Schema: workflow; Owner: -
--

CREATE TABLE workflow.event_destinations (
    id uuid NOT NULL,
    destination jsonb NOT NULL,
    "createdAt" timestamp(3) with time zone DEFAULT CURRENT_TIMESTAMP(3) NOT NULL,
    "updatedAt" timestamp(3) with time zone DEFAULT CURRENT_TIMESTAMP(3) NOT NULL
);


--
-- Name: execution_data; Type: TABLE; Schema: workflow; Owner: -
--

CREATE TABLE workflow.execution_data (
    "executionId" integer NOT NULL,
    "workflowData" json NOT NULL,
    data text NOT NULL
);


--
-- Name: execution_entity; Type: TABLE; Schema: workflow; Owner: -
--

CREATE TABLE workflow.execution_entity (
    id integer NOT NULL,
    finished boolean NOT NULL,
    mode character varying NOT NULL,
    "retryOf" character varying,
    "retrySuccessId" character varying,
    "startedAt" timestamp(3) with time zone NOT NULL,
    "stoppedAt" timestamp(3) with time zone,
    "waitTill" timestamp(3) with time zone,
    status character varying,
    "workflowId" character varying(36) NOT NULL,
    "deletedAt" timestamp(3) with time zone
);


--
-- Name: execution_entity_id_seq; Type: SEQUENCE; Schema: workflow; Owner: -
--

CREATE SEQUENCE workflow.execution_entity_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: execution_entity_id_seq; Type: SEQUENCE OWNED BY; Schema: workflow; Owner: -
--

ALTER SEQUENCE workflow.execution_entity_id_seq OWNED BY workflow.execution_entity.id;


--
-- Name: execution_metadata; Type: TABLE; Schema: workflow; Owner: -
--

CREATE TABLE workflow.execution_metadata (
    id integer NOT NULL,
    "executionId" integer NOT NULL,
    key text NOT NULL,
    value text NOT NULL
);


--
-- Name: execution_metadata_id_seq; Type: SEQUENCE; Schema: workflow; Owner: -
--

CREATE SEQUENCE workflow.execution_metadata_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: execution_metadata_id_seq; Type: SEQUENCE OWNED BY; Schema: workflow; Owner: -
--

ALTER SEQUENCE workflow.execution_metadata_id_seq OWNED BY workflow.execution_metadata.id;


--
-- Name: installed_nodes; Type: TABLE; Schema: workflow; Owner: -
--

CREATE TABLE workflow.installed_nodes (
    name character varying(200) NOT NULL,
    type character varying(200) NOT NULL,
    "latestVersion" integer DEFAULT 1 NOT NULL,
    package character varying(241) NOT NULL
);


--
-- Name: installed_packages; Type: TABLE; Schema: workflow; Owner: -
--

CREATE TABLE workflow.installed_packages (
    "packageName" character varying(214) NOT NULL,
    "installedVersion" character varying(50) NOT NULL,
    "authorName" character varying(70),
    "authorEmail" character varying(70),
    "createdAt" timestamp(3) with time zone DEFAULT CURRENT_TIMESTAMP(3) NOT NULL,
    "updatedAt" timestamp(3) with time zone DEFAULT CURRENT_TIMESTAMP(3) NOT NULL
);


--
-- Name: migrations; Type: TABLE; Schema: workflow; Owner: -
--

CREATE TABLE workflow.migrations (
    id integer NOT NULL,
    "timestamp" bigint NOT NULL,
    name character varying NOT NULL
);


--
-- Name: migrations_id_seq; Type: SEQUENCE; Schema: workflow; Owner: -
--

CREATE SEQUENCE workflow.migrations_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: migrations_id_seq; Type: SEQUENCE OWNED BY; Schema: workflow; Owner: -
--

ALTER SEQUENCE workflow.migrations_id_seq OWNED BY workflow.migrations.id;


--
-- Name: role; Type: TABLE; Schema: workflow; Owner: -
--

CREATE TABLE workflow.role (
    id integer NOT NULL,
    name character varying(32) NOT NULL,
    scope character varying(255) NOT NULL,
    "createdAt" timestamp(3) with time zone DEFAULT CURRENT_TIMESTAMP(3) NOT NULL,
    "updatedAt" timestamp(3) with time zone DEFAULT CURRENT_TIMESTAMP(3) NOT NULL
);


--
-- Name: role_id_seq; Type: SEQUENCE; Schema: workflow; Owner: -
--

CREATE SEQUENCE workflow.role_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: role_id_seq; Type: SEQUENCE OWNED BY; Schema: workflow; Owner: -
--

ALTER SEQUENCE workflow.role_id_seq OWNED BY workflow.role.id;


--
-- Name: settings; Type: TABLE; Schema: workflow; Owner: -
--

CREATE TABLE workflow.settings (
    key character varying(255) NOT NULL,
    value text NOT NULL,
    "loadOnStartup" boolean DEFAULT false NOT NULL
);


--
-- Name: shared_credentials; Type: TABLE; Schema: workflow; Owner: -
--

CREATE TABLE workflow.shared_credentials (
    "createdAt" timestamp(3) with time zone DEFAULT CURRENT_TIMESTAMP(3) NOT NULL,
    "updatedAt" timestamp(3) with time zone DEFAULT CURRENT_TIMESTAMP(3) NOT NULL,
    "roleId" integer NOT NULL,
    "userId" uuid NOT NULL,
    "credentialsId" character varying(36) NOT NULL
);


--
-- Name: shared_workflow; Type: TABLE; Schema: workflow; Owner: -
--

CREATE TABLE workflow.shared_workflow (
    "createdAt" timestamp(3) with time zone DEFAULT CURRENT_TIMESTAMP(3) NOT NULL,
    "updatedAt" timestamp(3) with time zone DEFAULT CURRENT_TIMESTAMP(3) NOT NULL,
    "roleId" integer NOT NULL,
    "userId" uuid NOT NULL,
    "workflowId" character varying(36) NOT NULL
);


--
-- Name: tag_entity; Type: TABLE; Schema: workflow; Owner: -
--

CREATE TABLE workflow.tag_entity (
    name character varying(24) NOT NULL,
    "createdAt" timestamp(3) with time zone DEFAULT CURRENT_TIMESTAMP(3) NOT NULL,
    "updatedAt" timestamp(3) with time zone DEFAULT CURRENT_TIMESTAMP(3) NOT NULL,
    id character varying(36) NOT NULL
);


--
-- Name: user; Type: TABLE; Schema: workflow; Owner: -
--

CREATE TABLE workflow."user" (
    id uuid DEFAULT uuid_in((OVERLAY(OVERLAY(md5((((random())::text || ':'::text) || (clock_timestamp())::text)) PLACING '4'::text FROM 13) PLACING to_hex((floor(((random() * (((11 - 8) + 1))::double precision) + (8)::double precision)))::integer) FROM 17))::cstring) NOT NULL,
    email character varying(255),
    "firstName" character varying(32),
    "lastName" character varying(32),
    password character varying(255),
    "personalizationAnswers" json,
    "createdAt" timestamp(3) with time zone DEFAULT CURRENT_TIMESTAMP(3) NOT NULL,
    "updatedAt" timestamp(3) with time zone DEFAULT CURRENT_TIMESTAMP(3) NOT NULL,
    "globalRoleId" integer NOT NULL,
    settings json,
    "apiKey" character varying(255),
    disabled boolean DEFAULT false NOT NULL,
    "mfaEnabled" boolean DEFAULT false NOT NULL,
    "mfaSecret" text,
    "mfaRecoveryCodes" text
);


--
-- Name: variables; Type: TABLE; Schema: workflow; Owner: -
--

CREATE TABLE workflow.variables (
    key character varying(50) NOT NULL,
    type character varying(50) DEFAULT 'string'::character varying NOT NULL,
    value character varying(255),
    id character varying(36) NOT NULL
);


--
-- Name: webhook_entity; Type: TABLE; Schema: workflow; Owner: -
--

CREATE TABLE workflow.webhook_entity (
    "webhookPath" character varying NOT NULL,
    method character varying NOT NULL,
    node character varying NOT NULL,
    "webhookId" character varying,
    "pathLength" integer,
    "workflowId" character varying(36) NOT NULL
);


--
-- Name: workflow_entity; Type: TABLE; Schema: workflow; Owner: -
--

CREATE TABLE workflow.workflow_entity (
    name character varying(128) NOT NULL,
    active boolean NOT NULL,
    nodes json NOT NULL,
    connections json NOT NULL,
    "createdAt" timestamp(3) with time zone DEFAULT CURRENT_TIMESTAMP(3) NOT NULL,
    "updatedAt" timestamp(3) with time zone DEFAULT CURRENT_TIMESTAMP(3) NOT NULL,
    settings json,
    "staticData" json,
    "pinData" json,
    "versionId" character(36),
    "triggerCount" integer DEFAULT 0 NOT NULL,
    id character varying(36) NOT NULL,
    meta json,
    "sugerOrgId" character varying(36) DEFAULT ''::character varying NOT NULL
);


--
-- Name: workflow_history; Type: TABLE; Schema: workflow; Owner: -
--

CREATE TABLE workflow.workflow_history (
    "versionId" character varying(36) NOT NULL,
    "workflowId" character varying(36) NOT NULL,
    authors character varying(255) NOT NULL,
    "createdAt" timestamp(3) with time zone DEFAULT CURRENT_TIMESTAMP(3) NOT NULL,
    "updatedAt" timestamp(3) with time zone DEFAULT CURRENT_TIMESTAMP(3) NOT NULL,
    nodes json NOT NULL,
    connections json NOT NULL
);


--
-- Name: workflow_statistics; Type: TABLE; Schema: workflow; Owner: -
--

CREATE TABLE workflow.workflow_statistics (
    count integer DEFAULT 0,
    "latestEvent" timestamp(3) with time zone,
    name character varying(128) NOT NULL,
    "workflowId" character varying(36) NOT NULL
);

--
-- Name: workflows_tags; Type: TABLE; Schema: workflow; Owner: -
--

CREATE TABLE workflow.workflows_tags (
    "workflowId" character varying(36) NOT NULL,
    "tagId" character varying(36) NOT NULL
);


--
-- Name: auth_provider_sync_history id; Type: DEFAULT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.auth_provider_sync_history ALTER COLUMN id SET DEFAULT nextval('workflow.auth_provider_sync_history_id_seq'::regclass);


--
-- Name: execution_entity id; Type: DEFAULT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.execution_entity ALTER COLUMN id SET DEFAULT nextval('workflow.execution_entity_id_seq'::regclass);


--
-- Name: execution_metadata id; Type: DEFAULT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.execution_metadata ALTER COLUMN id SET DEFAULT nextval('workflow.execution_metadata_id_seq'::regclass);


--
-- Name: migrations id; Type: DEFAULT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.migrations ALTER COLUMN id SET DEFAULT nextval('workflow.migrations_id_seq'::regclass);


--
-- Name: role id; Type: DEFAULT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.role ALTER COLUMN id SET DEFAULT nextval('workflow.role_id_seq'::regclass);

--
-- Name: integration integration_pkey; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.integration
    ADD CONSTRAINT integration_pkey PRIMARY KEY (organization_id, partner, service);


--
-- Name: organization organization_pkey; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.organization
    ADD CONSTRAINT organization_pkey PRIMARY KEY (id);


--
-- Name: role role_pkey; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.role
    ADD CONSTRAINT role_pkey PRIMARY KEY (organization_id, id);


--
-- Name: user user_email_unique; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity."user"
    ADD CONSTRAINT user_email_unique UNIQUE (email);


--
-- Name: user_organization user_organization_pkey; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity.user_organization
    ADD CONSTRAINT user_organization_pkey PRIMARY KEY (user_id, organization_id);


--
-- Name: user user_pkey; Type: CONSTRAINT; Schema: identity; Owner: -
--

ALTER TABLE ONLY identity."user"
    ADD CONSTRAINT user_pkey PRIMARY KEY (id);

--
-- Name: webhook webhook_pkey; Type: CONSTRAINT; Schema: notification; Owner: -
--

ALTER TABLE ONLY notification.webhook
    ADD CONSTRAINT webhook_pkey PRIMARY KEY (organization_id, id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: installed_packages PK_08cc9197c39b028c1e9beca225940576fd1a5804; Type: CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.installed_packages
    ADD CONSTRAINT "PK_08cc9197c39b028c1e9beca225940576fd1a5804" PRIMARY KEY ("packageName");


--
-- Name: migrations PK_8c82d7f526340ab734260ea46be; Type: CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.migrations
    ADD CONSTRAINT "PK_8c82d7f526340ab734260ea46be" PRIMARY KEY (id);


--
-- Name: installed_nodes PK_8ebd28194e4f792f96b5933423fc439df97d9689; Type: CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.installed_nodes
    ADD CONSTRAINT "PK_8ebd28194e4f792f96b5933423fc439df97d9689" PRIMARY KEY (name);


--
-- Name: webhook_entity PK_b21ace2e13596ccd87dc9bf4ea6; Type: CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.webhook_entity
    ADD CONSTRAINT "PK_b21ace2e13596ccd87dc9bf4ea6" PRIMARY KEY ("webhookPath", method);


--
-- Name: workflow_history PK_b6572dd6173e4cd06fe79937b58; Type: CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.workflow_history
    ADD CONSTRAINT "PK_b6572dd6173e4cd06fe79937b58" PRIMARY KEY ("versionId");


--
-- Name: settings PK_dc0fe14e6d9943f268e7b119f69ab8bd; Type: CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.settings
    ADD CONSTRAINT "PK_dc0fe14e6d9943f268e7b119f69ab8bd" PRIMARY KEY (key);


--
-- Name: role PK_e853ce24e8200abe5721d2c6ac552b73; Type: CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.role
    ADD CONSTRAINT "PK_e853ce24e8200abe5721d2c6ac552b73" PRIMARY KEY (id);


--
-- Name: user PK_ea8f538c94b6e352418254ed6474a81f; Type: CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow."user"
    ADD CONSTRAINT "PK_ea8f538c94b6e352418254ed6474a81f" PRIMARY KEY (id);


--
-- Name: role UQ_5b49d0f504f7ef31045a1fb2eb8; Type: CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.role
    ADD CONSTRAINT "UQ_5b49d0f504f7ef31045a1fb2eb8" UNIQUE (scope, name);


--
-- Name: user UQ_e12875dfb3b1d92d7d7c5377e2; Type: CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow."user"
    ADD CONSTRAINT "UQ_e12875dfb3b1d92d7d7c5377e2" UNIQUE (email);


--
-- Name: auth_identity auth_identity_pkey; Type: CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.auth_identity
    ADD CONSTRAINT auth_identity_pkey PRIMARY KEY ("providerId", "providerType");


--
-- Name: auth_provider_sync_history auth_provider_sync_history_pkey; Type: CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.auth_provider_sync_history
    ADD CONSTRAINT auth_provider_sync_history_pkey PRIMARY KEY (id);


--
-- Name: credentials_entity credentials_entity_pkey; Type: CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.credentials_entity
    ADD CONSTRAINT credentials_entity_pkey PRIMARY KEY (id);


--
-- Name: event_destinations event_destinations_pkey; Type: CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.event_destinations
    ADD CONSTRAINT event_destinations_pkey PRIMARY KEY (id);


--
-- Name: execution_data execution_data_pkey; Type: CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.execution_data
    ADD CONSTRAINT execution_data_pkey PRIMARY KEY ("executionId");


--
-- Name: execution_metadata execution_metadata_pkey; Type: CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.execution_metadata
    ADD CONSTRAINT execution_metadata_pkey PRIMARY KEY (id);


--
-- Name: execution_entity pk_e3e63bbf986767844bbe1166d4e; Type: CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.execution_entity
    ADD CONSTRAINT pk_e3e63bbf986767844bbe1166d4e PRIMARY KEY (id);


--
-- Name: shared_credentials pk_shared_credentials_id; Type: CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.shared_credentials
    ADD CONSTRAINT pk_shared_credentials_id PRIMARY KEY ("userId", "credentialsId");


--
-- Name: shared_workflow pk_shared_workflow_id; Type: CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.shared_workflow
    ADD CONSTRAINT pk_shared_workflow_id PRIMARY KEY ("userId", "workflowId");


--
-- Name: workflow_statistics pk_workflow_statistics; Type: CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.workflow_statistics
    ADD CONSTRAINT pk_workflow_statistics PRIMARY KEY ("workflowId", name);


--
-- Name: workflows_tags pk_workflows_tags; Type: CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.workflows_tags
    ADD CONSTRAINT pk_workflows_tags PRIMARY KEY ("workflowId", "tagId");


--
-- Name: tag_entity tag_entity_pkey; Type: CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.tag_entity
    ADD CONSTRAINT tag_entity_pkey PRIMARY KEY (id);


--
-- Name: variables variables_key_key; Type: CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.variables
    ADD CONSTRAINT variables_key_key UNIQUE (key);


--
-- Name: variables variables_pkey; Type: CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.variables
    ADD CONSTRAINT variables_pkey PRIMARY KEY (id);


--
-- Name: workflow_entity workflow_entity_pkey; Type: CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.workflow_entity
    ADD CONSTRAINT workflow_entity_pkey PRIMARY KEY (id);


--
-- Name: IDX_1e31657f5fe46816c34be7c1b4; Type: INDEX; Schema: workflow; Owner: -
--

CREATE INDEX "IDX_1e31657f5fe46816c34be7c1b4" ON workflow.workflow_history USING btree ("workflowId");


--
-- Name: IDX_6d44376da6c1058b5e81ed8a154e1fee106046eb; Type: INDEX; Schema: workflow; Owner: -
--

CREATE INDEX "IDX_6d44376da6c1058b5e81ed8a154e1fee106046eb" ON workflow.execution_metadata USING btree ("executionId");


--
-- Name: IDX_85b981df7b444f905f8bf50747; Type: INDEX; Schema: workflow; Owner: -
--

CREATE INDEX "IDX_85b981df7b444f905f8bf50747" ON workflow.execution_entity USING btree ("waitTill", id);


--
-- Name: IDX_execution_entity_deletedAt; Type: INDEX; Schema: workflow; Owner: -
--

CREATE INDEX "IDX_execution_entity_deletedAt" ON workflow.execution_entity USING btree ("deletedAt");


--
-- Name: IDX_execution_entity_stoppedAt; Type: INDEX; Schema: workflow; Owner: -
--

CREATE INDEX "IDX_execution_entity_stoppedAt" ON workflow.execution_entity USING btree ("stoppedAt");


--
-- Name: IDX_workflow_entity_name; Type: INDEX; Schema: workflow; Owner: -
--

CREATE INDEX "IDX_workflow_entity_name" ON workflow.workflow_entity USING btree (name);


--
-- Name: UQ_ie0zomxves9w3p774drfrkxtj5; Type: INDEX; Schema: workflow; Owner: -
--

CREATE UNIQUE INDEX "UQ_ie0zomxves9w3p774drfrkxtj5" ON workflow."user" USING btree ("apiKey");


--
-- Name: idx_07fde106c0b471d8cc80a64fc8; Type: INDEX; Schema: workflow; Owner: -
--

CREATE INDEX idx_07fde106c0b471d8cc80a64fc8 ON workflow.credentials_entity USING btree (type);


--
-- Name: idx_16f4436789e804e3e1c9eeb240; Type: INDEX; Schema: workflow; Owner: -
--

CREATE INDEX idx_16f4436789e804e3e1c9eeb240 ON workflow.webhook_entity USING btree ("webhookId", method, "pathLength");


--
-- Name: idx_812eb05f7451ca757fb98444ce; Type: INDEX; Schema: workflow; Owner: -
--

CREATE UNIQUE INDEX idx_812eb05f7451ca757fb98444ce ON workflow.tag_entity USING btree (name);


--
-- Name: idx_execution_entity_workflow_id_id; Type: INDEX; Schema: workflow; Owner: -
--

CREATE INDEX idx_execution_entity_workflow_id_id ON workflow.execution_entity USING btree ("workflowId", id);


--
-- Name: idx_shared_credentials_credentials_id; Type: INDEX; Schema: workflow; Owner: -
--

CREATE INDEX idx_shared_credentials_credentials_id ON workflow.shared_credentials USING btree ("credentialsId");


--
-- Name: idx_shared_workflow_workflow_id; Type: INDEX; Schema: workflow; Owner: -
--

CREATE INDEX idx_shared_workflow_workflow_id ON workflow.shared_workflow USING btree ("workflowId");


--
-- Name: idx_workflows_tags_workflow_id; Type: INDEX; Schema: workflow; Owner: -
--

CREATE INDEX idx_workflows_tags_workflow_id ON workflow.workflows_tags USING btree ("workflowId");


--
-- Name: pk_credentials_entity_id; Type: INDEX; Schema: workflow; Owner: -
--

CREATE UNIQUE INDEX pk_credentials_entity_id ON workflow.credentials_entity USING btree (id);


--
-- Name: pk_tag_entity_id; Type: INDEX; Schema: workflow; Owner: -
--

CREATE UNIQUE INDEX pk_tag_entity_id ON workflow.tag_entity USING btree (id);


--
-- Name: pk_variables_id; Type: INDEX; Schema: workflow; Owner: -
--

CREATE UNIQUE INDEX pk_variables_id ON workflow.variables USING btree (id);


--
-- Name: pk_workflow_entity_id; Type: INDEX; Schema: workflow; Owner: -
--

CREATE UNIQUE INDEX pk_workflow_entity_id ON workflow.workflow_entity USING btree (id);


--
-- Name: ticket ticket_organization_id_fkey; Type: FK CONSTRAINT; Schema: support; Owner: -
--

ALTER TABLE ONLY support.ticket
    ADD CONSTRAINT ticket_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES identity.organization(id);


--
-- Name: workflow_history FK_1e31657f5fe46816c34be7c1b4b; Type: FK CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.workflow_history
    ADD CONSTRAINT "FK_1e31657f5fe46816c34be7c1b4b" FOREIGN KEY ("workflowId") REFERENCES workflow.workflow_entity(id) ON DELETE CASCADE;


--
-- Name: shared_workflow FK_3540da03964527aa24ae014b780; Type: FK CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.shared_workflow
    ADD CONSTRAINT "FK_3540da03964527aa24ae014b780" FOREIGN KEY ("roleId") REFERENCES workflow.role(id);


--
-- Name: shared_credentials FK_484f0327e778648dd04f1d70493; Type: FK CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.shared_credentials
    ADD CONSTRAINT "FK_484f0327e778648dd04f1d70493" FOREIGN KEY ("userId") REFERENCES workflow."user"(id) ON DELETE CASCADE;


--
-- Name: installed_nodes FK_73f857fc5dce682cef8a99c11dbddbc969618951; Type: FK CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.installed_nodes
    ADD CONSTRAINT "FK_73f857fc5dce682cef8a99c11dbddbc969618951" FOREIGN KEY (package) REFERENCES workflow.installed_packages("packageName") ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: shared_workflow FK_82b2fd9ec4e3e24209af8160282; Type: FK CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.shared_workflow
    ADD CONSTRAINT "FK_82b2fd9ec4e3e24209af8160282" FOREIGN KEY ("userId") REFERENCES workflow."user"(id) ON DELETE CASCADE;


--
-- Name: shared_credentials FK_c68e056637562000b68f480815a; Type: FK CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.shared_credentials
    ADD CONSTRAINT "FK_c68e056637562000b68f480815a" FOREIGN KEY ("roleId") REFERENCES workflow.role(id);


--
-- Name: user FK_f0609be844f9200ff4365b1bb3d; Type: FK CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow."user"
    ADD CONSTRAINT "FK_f0609be844f9200ff4365b1bb3d" FOREIGN KEY ("globalRoleId") REFERENCES workflow.role(id);


--
-- Name: auth_identity auth_identity_userId_fkey; Type: FK CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.auth_identity
    ADD CONSTRAINT "auth_identity_userId_fkey" FOREIGN KEY ("userId") REFERENCES workflow."user"(id);


--
-- Name: execution_data execution_data_fk; Type: FK CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.execution_data
    ADD CONSTRAINT execution_data_fk FOREIGN KEY ("executionId") REFERENCES workflow.execution_entity(id) ON DELETE CASCADE;


--
-- Name: execution_metadata execution_metadata_fk; Type: FK CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.execution_metadata
    ADD CONSTRAINT execution_metadata_fk FOREIGN KEY ("executionId") REFERENCES workflow.execution_entity(id) ON DELETE CASCADE;


--
-- Name: execution_entity fk_execution_entity_workflow_id; Type: FK CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.execution_entity
    ADD CONSTRAINT fk_execution_entity_workflow_id FOREIGN KEY ("workflowId") REFERENCES workflow.workflow_entity(id) ON DELETE CASCADE;


--
-- Name: shared_credentials fk_shared_credentials_credentials_id; Type: FK CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.shared_credentials
    ADD CONSTRAINT fk_shared_credentials_credentials_id FOREIGN KEY ("credentialsId") REFERENCES workflow.credentials_entity(id) ON DELETE CASCADE;


--
-- Name: shared_workflow fk_shared_workflow_workflow_id; Type: FK CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.shared_workflow
    ADD CONSTRAINT fk_shared_workflow_workflow_id FOREIGN KEY ("workflowId") REFERENCES workflow.workflow_entity(id) ON DELETE CASCADE;


--
-- Name: webhook_entity fk_webhook_entity_workflow_id; Type: FK CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.webhook_entity
    ADD CONSTRAINT fk_webhook_entity_workflow_id FOREIGN KEY ("workflowId") REFERENCES workflow.workflow_entity(id) ON DELETE CASCADE;


--
-- Name: workflow_statistics fk_workflow_statistics_workflow_id; Type: FK CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.workflow_statistics
    ADD CONSTRAINT fk_workflow_statistics_workflow_id FOREIGN KEY ("workflowId") REFERENCES workflow.workflow_entity(id) ON DELETE CASCADE;


--
-- Name: workflows_tags fk_workflows_tags_tag_id; Type: FK CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.workflows_tags
    ADD CONSTRAINT fk_workflows_tags_tag_id FOREIGN KEY ("tagId") REFERENCES workflow.tag_entity(id) ON DELETE CASCADE;


--
-- Name: workflows_tags fk_workflows_tags_workflow_id; Type: FK CONSTRAINT; Schema: workflow; Owner: -
--

ALTER TABLE ONLY workflow.workflows_tags
    ADD CONSTRAINT fk_workflows_tags_workflow_id FOREIGN KEY ("workflowId") REFERENCES workflow.workflow_entity(id) ON DELETE CASCADE;