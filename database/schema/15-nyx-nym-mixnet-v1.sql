CREATE TABLE nyx_nym_mixnet_v1_mixnode
(
    identity_key           TEXT             UNIQUE PRIMARY KEY,
    is_bonded              BOOLEAN          NOT NULL,

    -- values: in_active_set, in_standby_set, inactive
    last_mixnet_status     TEXT             NULL
);
CREATE INDEX nyx_nym_mixnet_v1_mixnode_status_index ON nyx_nym_mixnet_v1_mixnode (last_mixnet_status);

CREATE TABLE nyx_nym_mixnet_v1_mixnode_status
(
    -- values: in_active_set, in_standby_set, inactive
    mixnet_status           TEXT            NOT NULL,

    -- in the range 0 to 100
    routing_score           INTEGER         NOT NULL,

    identity_key            TEXT            NOT NULL REFERENCES nyx_nym_mixnet_v1_mixnode (identity_key),
    executed_at             TIMESTAMP       NOT NULL,
    height                  BIGINT          NOT NULL REFERENCES block (height),
    hash                    TEXT            NOT NULL
);
CREATE INDEX nyx_nym_mixnet_v1_mixnode_status_height_index ON nyx_nym_mixnet_v1_mixnode_status (height);

CREATE TABLE nyx_nym_mixnet_v1_mixnode_events
(
    -- values: bond, unbond, delegate, undelegate, claim, compound
    event_kind              TEXT            NOT NULL,
    -- values: mixnode_operator, mixnode_delegator, mixnet_rewarding, mixnet_monitoring
    actor                   TEXT            NOT NULL,
    sender                  TEXT            NOT NULL,
    proxy                   TEXT            NULL,
    identity_key            TEXT            NOT NULL REFERENCES nyx_nym_mixnet_v1_mixnode (identity_key),
    amount                  COIN[]          NOT NULL DEFAULT '{}',
    fee                     COIN[]          NOT NULL DEFAULT '{}',
    contract_address        TEXT            NOT NULL REFERENCES wasm_contract (contract_address),
    event_type              TEXT            NULL,
    attributes              JSONB           NOT NULL DEFAULT '{}'::JSONB,
    executed_at             TIMESTAMP       NOT NULL,
    height                  BIGINT          NOT NULL REFERENCES block (height),
    hash                    TEXT            NOT NULL
);
CREATE INDEX nyx_nym_mixnet_v1_mixnode_events_height_index ON nyx_nym_mixnet_v1_mixnode_events (height);

CREATE TABLE nyx_nym_mixnet_v1_mixnode_reward
(
    sender                     TEXT            NOT NULL,
    identity_key               TEXT            NOT NULL REFERENCES nyx_nym_mixnet_v1_mixnode (identity_key),
    total_node_reward          COIN[]          NOT NULL DEFAULT '{}',
    total_delegations          COIN[]          NOT NULL DEFAULT '{}',
    operator_reward            COIN[]          NOT NULL DEFAULT '{}',
    unit_delegator_reward      BIGINT          NOT NULL,
    apy                        FLOAT           NOT NULL,
    staking_supply             COIN[]          NOT NULL DEFAULT '{}',
    profit_margin_percentage   INTEGER         NOT NULL,
    contract_address           TEXT            NOT NULL REFERENCES wasm_contract (contract_address),
    event_type                 TEXT            NULL,
    attributes                 JSONB           NOT NULL DEFAULT '{}'::JSONB,
    executed_at                TIMESTAMP       NOT NULL,
    height                     BIGINT          NOT NULL REFERENCES block (height),
    hash                       TEXT            NOT NULL,
    UNIQUE (sender, contract_address, identity_key, height, hash)
);
CREATE INDEX nyx_nym_mixnet_v1_mixnode_reward_height_index ON nyx_nym_mixnet_v1_mixnode_reward (height);

