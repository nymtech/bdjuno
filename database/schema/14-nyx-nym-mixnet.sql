CREATE TABLE nyx_nym_mixnet_v1_gateway
(
    identity_key           TEXT             UNIQUE PRIMARY KEY,
    is_bonded              BOOLEAN          NOT NULL
);

CREATE TABLE nyx_nym_mixnet_v1_gateway_events
(
    -- values: bond, unbond
    event_kind              TEXT            NOT NULL,
    sender                  TEXT            NOT NULL,
    proxy                   TEXT            NULL,
    identity_key            TEXT            NOT NULL REFERENCES nyx_nym_mixnet_v1_gateway (identity_key),
    amount                  COIN            NULL,
    fee                     COIN            NULL,
    contract_address        TEXT            NOT NULL REFERENCES wasm_contract (contract_address),
    event_type              TEXT            NULL,
    attributes              JSONB           NOT NULL DEFAULT '{}'::JSONB,
    executed_at             TIMESTAMP       NOT NULL,
    height                  BIGINT          NOT NULL REFERENCES block (height),
    hash                    TEXT            NOT NULL
);
CREATE INDEX nyx_nym_mixnet_v1_gateway_events_height_index ON nyx_nym_mixnet_v1_gateway_events (height);
