CREATE TABLE nyx_nym_mixnet_v1_mixnode
(
    identity_key          TEXT UNIQUE PRIMARY KEY,
    last_is_bonded_status BOOLEAN NULL,

    -- values: in_active_set, in_standby_set, inactive
    last_mixnet_status    TEXT    NULL
);
CREATE INDEX nyx_nym_mixnet_v1_mixnode_status_index ON nyx_nym_mixnet_v1_mixnode (last_mixnet_status);

CREATE TABLE nyx_nym_mixnet_v1_mixnode_status
(
    -- values: in_active_set, in_standby_set, inactive
    mixnet_status TEXT      NOT NULL,

    -- in the range 0 to 100
    routing_score INTEGER   NOT NULL,

    identity_key  TEXT      NOT NULL REFERENCES nyx_nym_mixnet_v1_mixnode (identity_key),
    executed_at   TIMESTAMP NOT NULL,
    height        BIGINT    NOT NULL REFERENCES block (height),
    hash          TEXT      NOT NULL
);
CREATE INDEX nyx_nm_v1_m_status_height_index ON nyx_nym_mixnet_v1_mixnode_status (height);
CREATE INDEX nyx_nm_v1_m_identity_key_index ON nyx_nym_mixnet_v1_mixnode_status (identity_key);

CREATE TABLE nyx_nym_mixnet_v1_mixnode_events
(
    -- values: bond, unbond, delegate, undelegate, claim
    event_kind    TEXT      NOT NULL,
    -- values: mixnode_operator, mixnode_delegator, mixnet_rewarding, mixnet_monitoring
    actor         TEXT      NOT NULL,
    sender        TEXT      NOT NULL,
    proxy         TEXT      NULL,
    identity_key  TEXT      NOT NULL REFERENCES nyx_nym_mixnet_v1_mixnode (identity_key),
    executed_at   TIMESTAMP NOT NULL,
    height        BIGINT    NOT NULL REFERENCES block (height),
    hash          TEXT      NOT NULL,
    message_index BIGINT    NOT NULL
);
CREATE INDEX nyx_nm_v1_me_height_index ON nyx_nym_mixnet_v1_mixnode_events (identity_key, height);
CREATE INDEX nyx_nm_v1_me_identity_key_executed_at_index ON nyx_nym_mixnet_v1_mixnode_events (identity_key, executed_at);

CREATE TABLE nyx_nym_mixnet_v1_mixnode_reward
(
    identity_key          TEXT      NOT NULL REFERENCES nyx_nym_mixnet_v1_mixnode (identity_key),
    total_node_reward     COIN[]    NOT NULL DEFAULT '{}',
    total_delegations     COIN[]    NOT NULL DEFAULT '{}',
    operator_reward       COIN[]    NOT NULL DEFAULT '{}',
    unit_delegator_reward BIGINT    NOT NULL,
    apy                   FLOAT     NOT NULL,
    executed_at           TIMESTAMP NOT NULL,
    height                BIGINT    NOT NULL REFERENCES block (height),
    hash                  TEXT      NOT NULL,
    message_index         BIGINT    NOT NULL,
    UNIQUE (identity_key, height, hash, message_index)
);
CREATE INDEX nyx_nm_v1_mr_height_index ON nyx_nym_mixnet_v1_mixnode_reward (height);
CREATE INDEX nyx_nm_v1_mr_identity_key_index ON nyx_nym_mixnet_v1_mixnode_reward (identity_key);

CREATE TABLE nyx_nym_mixnet_v1_gateway
(
    identity_key          TEXT UNIQUE PRIMARY KEY,
    last_is_bonded_status BOOLEAN NULL
);

CREATE TABLE nyx_nym_mixnet_v1_gateway_events
(
    -- values: bond, unbond
    event_kind    TEXT      NOT NULL,
    sender        TEXT      NOT NULL,
    proxy         TEXT      NULL,
    identity_key  TEXT      NOT NULL REFERENCES nyx_nym_mixnet_v1_gateway (identity_key),
    executed_at   TIMESTAMP NOT NULL,
    height        BIGINT    NOT NULL REFERENCES block (height),
    hash          TEXT      NOT NULL,
    message_index BIGINT    NOT NULL
);
CREATE INDEX nyx_nm_v1_ge_height_index ON nyx_nym_mixnet_v1_gateway_events (identity_key, height);
CREATE INDEX nyx_nm_v1_ge_identity_key_executed_at_index ON nyx_nym_mixnet_v1_gateway_events (identity_key, executed_at);


