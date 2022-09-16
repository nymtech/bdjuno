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
