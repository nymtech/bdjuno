CREATE TYPE ACCESS_CONFIG AS
(
    permission  INT,
    address     TEXT
);

CREATE TABLE wasm_params
(
    one_row_id                      BOOLEAN         NOT NULL DEFAULT TRUE PRIMARY KEY,
    code_upload_access              ACCESS_CONFIG   NOT NULL,
    instantiate_default_permission  INT             NOT NULL,
    height                          BIGINT          NOT NULL
);


CREATE TABLE wasm_code
(
    sender                  TEXT            NULL,
    byte_code               BYTEA           NOT NULL,
    instantiate_permission  ACCESS_CONFIG   NULL,
    code_id                 BIGINT          NOT NULL UNIQUE,
    height                  BIGINT          NOT NULL
);
CREATE INDEX wasm_code_height_index ON wasm_code (height);

CREATE TABLE wasm_contract
(
    sender                  TEXT            NULL,
    creator                 TEXT            NOT NULL REFERENCES account (address),
    admin                   TEXT            NULL,
    code_id                 BIGINT          NOT NULL REFERENCES wasm_code (code_id),
    label                   TEXT            NULL,
    raw_contract_message    JSONB           NOT NULL DEFAULT '{}'::JSONB,
    funds                   COIN[]          NOT NULL DEFAULT '{}',
    contract_address        TEXT            NOT NULL UNIQUE,
    data                    TEXT            NULL,
    instantiated_at         TIMESTAMP       NOT NULL,
    contract_info_extension TEXT            NULL,
    contract_states         JSONB           NOT NULL DEFAULT '{}'::JSONB,
    height                  BIGINT          NOT NULL
);
CREATE INDEX wasm_contract_height_index ON wasm_contract (height);
CREATE INDEX wasm_contract_creator_index ON wasm_contract (creator);
CREATE INDEX wasm_contract_label_index ON wasm_contract (label);

CREATE TABLE wasm_execute_contract
(
    sender                  TEXT            NOT NULL,
    contract_address        TEXT            NOT NULL REFERENCES wasm_contract (contract_address),
    raw_contract_message    JSONB           NOT NULL DEFAULT '{}'::JSONB,
    funds                   COIN[]          NOT NULL DEFAULT '{}',
    message_type            TEXT            NULL,
    data                    TEXT            NULL,
    executed_at             TIMESTAMP       NOT NULL,
    height                  BIGINT          NOT NULL,
    hash                    TEXT            NOT NULL
);
CREATE INDEX execute_contract_height_index ON wasm_execute_contract (height);
CREATE INDEX execute_contract_executed_at_index ON wasm_execute_contract (executed_at);
CREATE INDEX execute_contract_message_type_index ON wasm_execute_contract (message_type);

CREATE TABLE wasm_execute_contract_event_types
(
    contract_address         TEXT   NOT NULL REFERENCES wasm_contract (contract_address),
    event_type               TEXT   NOT NULL,

    first_seen_height        BIGINT NOT NULL REFERENCES block (height),
    first_seen_hash          TEXT   NOT NULL,

    last_seen_height         BIGINT NOT NULL REFERENCES block (height),
    last_seen_hash           TEXT   NOT NULL,
    UNIQUE (contract_address, event_type)
);
CREATE INDEX wasm_execute_contract_event_types_index ON wasm_execute_contract_event_types (contract_address, event_type);
