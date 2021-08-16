CREATE TABLE tenants (
    -- base
    id BIGSERIAL PRIMARY KEY,
    uuid VARCHAR (36) NOT NULL,
    name VARCHAR (255) NULL,
    state SMALLINT NOT NULL,
    timezone VARCHAR (63) NOT NULL DEFAULT 'utc',
    created_time TIMESTAMPTZ NOT NULL DEFAULT (now() AT TIME ZONE 'utc'),
    modified_time TIMESTAMPTZ NOT NULL DEFAULT (now() AT TIME ZONE 'utc')
);
CREATE UNIQUE INDEX idx_tenant_uuid
ON tenants (uuid);

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    uuid VARCHAR (36) NOT NULL,
    tenant_id BIGINT NOT NULL,
    first_name VARCHAR (50) NULL,
    last_name VARCHAR (50) NULL,
    email VARCHAR (255) NOT NULL,
    password_algorithm VARCHAR (63) NOT NULL,
    password_hash VARCHAR (1027) NOT NULL,
    state SMALLINT NOT NULL DEFAULT 0,
    role_id SMALLINT NOT NULL DEFAULT 0,
    timezone VARCHAR (63) NOT NULL DEFAULT 'utc',
    created_time TIMESTAMPTZ NOT NULL DEFAULT (now() AT TIME ZONE 'utc'),
    modified_time TIMESTAMPTZ NOT NULL DEFAULT (now() AT TIME ZONE 'utc'),
    salt VARCHAR (127) NOT NULL DEFAULT '',
    was_email_activated BOOLEAN NOT NULL DEFAULT FALSE,
    pr_access_code VARCHAR (127) NOT NULL DEFAULT '',
    pr_expiry_time TIMESTAMPTZ NOT NULL DEFAULT (now() AT TIME ZONE 'utc'),
    FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);
CREATE UNIQUE INDEX idx_user_uuid
ON users (uuid);
CREATE UNIQUE INDEX idx_user_email
ON users (email);
CREATE INDEX idx_user_tenant_id
ON users (tenant_id);
