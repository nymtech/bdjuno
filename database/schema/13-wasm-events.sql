CREATE TABLE wasm_execute_contract_event
(
    sender                  TEXT            NOT NULL,
    contract_address        TEXT            NOT NULL REFERENCES wasm_contract (contract_address),
    event_type              TEXT            NULL,
    attributes              JSONB           NOT NULL DEFAULT '{}'::JSONB,
    executed_at             TIMESTAMP       NOT NULL,
    height                  BIGINT          NOT NULL REFERENCES block (height),
    hash                    TEXT            NOT NULL
);
CREATE INDEX wasm_execute_contract_event_height_index ON wasm_execute_contract_event (height);
CREATE INDEX wasm_execute_contract_event_hash_index ON wasm_execute_contract_event (hash);
CREATE INDEX wasm_execute_contract_event_event_type_index ON wasm_execute_contract_event (event_type);
