CREATE TABLE nyx_nym_mixnet_v2_mixnode
(
    mix_id                 BIGINT           UNIQUE PRIMARY KEY,
    identity_key           TEXT             NOT NULL,
    is_bonded              BOOLEAN          NOT NULL,

    -- values: in_active_set, in_standby_set, inactive
    last_mixnet_status     TEXT             NULL
);
CREATE INDEX nyx_nym_mixnet_v2_mixnode_status_index ON nyx_nym_mixnet_v2_mixnode (last_mixnet_status);
CREATE INDEX nyx_nym_mixnet_v2_mixnode_identity_key_index ON nyx_nym_mixnet_v2_mixnode (identity_key);

CREATE TABLE nyx_nym_mixnet_v2_mixnode_status
(
    -- values: in_active_set, in_standby_set, inactive
    mixnet_status           TEXT            NOT NULL,

    -- in the range 0 to 1
    routing_score           DECIMAL         NOT NULL,

    mix_id                  BIGINT          NOT NULL REFERENCES nyx_nym_mixnet_v2_mixnode (mix_id),
    executed_at             TIMESTAMP       NOT NULL,
    height                  BIGINT          NOT NULL REFERENCES block (height),
    hash                    TEXT            NOT NULL
);
CREATE INDEX nyx_nym_mixnet_v2_mixnode_status_height_index ON nyx_nym_mixnet_v2_mixnode_status (height);

CREATE TABLE nyx_nym_mixnet_v2_events
(
    -- values: bond, unbond, delegate, undelegate, claim
    event_kind              TEXT            NOT NULL,
    -- values: mixnode_operator, mixnode_delegator, mixnet_rewarding, mixnet_monitoring
    actor                   TEXT            NOT NULL,
    sender                  TEXT            NOT NULL,
    proxy                   TEXT            NULL,
    mix_id                  BIGINT          NOT NULL REFERENCES nyx_nym_mixnet_v2_mixnode (mix_id),
    identity_key            TEXT            NOT NULL,
    amount                  COIN            NULL,
    fee                     COIN            NULL,
    contract_address        TEXT            NOT NULL REFERENCES wasm_contract (contract_address),
    event_type              TEXT            NULL,
    attributes              JSONB           NOT NULL DEFAULT '{}'::JSONB,
    executed_at             TIMESTAMP       NOT NULL,
    height                  BIGINT          NOT NULL REFERENCES block (height),
    hash                    TEXT            NOT NULL
);
CREATE INDEX nyx_nym_mixnet_v2_events_height_index ON nyx_nym_mixnet_v2_events (height);

CREATE TABLE nyx_nym_mixnet_v2_mixnode_reward
(
    sender                     TEXT            NOT NULL,
    mix_id                     BIGINT          NOT NULL REFERENCES nyx_nym_mixnet_v2_mixnode (mix_id),
    operator_reward            COIN[]          NOT NULL DEFAULT '{}',
    delegates_reward           COIN[]          NOT NULL DEFAULT '{}',
    prior_delegates            COIN[]          NOT NULL DEFAULT '{}',
    prior_unit_delegation      BIGINT          NOT NULL,
    apy                        FLOAT           NOT NULL,
    contract_address           TEXT            NOT NULL REFERENCES wasm_contract (contract_address),
    event_type                 TEXT            NULL,
    attributes                 JSONB           NOT NULL DEFAULT '{}'::JSONB,
    executed_at                TIMESTAMP       NOT NULL,
    height                     BIGINT          NOT NULL REFERENCES block (height),
    hash                       TEXT            NOT NULL,
    UNIQUE (sender, contract_address, mix_id, height, hash)
);
CREATE INDEX nyx_nym_mixnet_v2_mixnode_reward_height_index ON nyx_nym_mixnet_v2_mixnode_reward (height);