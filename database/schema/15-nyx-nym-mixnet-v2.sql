CREATE TABLE nyx_nym_mixnet_v2_mixnode
(
    mix_id                BIGINT UNIQUE PRIMARY KEY,

    -- the identity key may be null, because a tx/event with only the mix_id might be processed out of order
    identity_key          TEXT    NULL,

    last_is_bonded_status BOOLEAN NULL,

    -- values: in_active_set, in_standby_set, inactive
    last_mixnet_status    TEXT    NULL
);
CREATE INDEX nyx_nm_v2_m_status_index ON nyx_nym_mixnet_v2_mixnode (last_mixnet_status);
CREATE INDEX nyx_nm_v2_m_identity_key_index ON nyx_nym_mixnet_v2_mixnode (identity_key);

CREATE TABLE nyx_nym_mixnet_v2_mixnode_status
(
    -- values: in_active_set, in_standby_set, inactive
    mixnet_status TEXT      NOT NULL,

    -- in the range 0 to 1
    routing_score DECIMAL   NOT NULL,

    mix_id        BIGINT    NOT NULL REFERENCES nyx_nym_mixnet_v2_mixnode (mix_id),
    executed_at   TIMESTAMP NOT NULL,
    height        BIGINT    NOT NULL REFERENCES block (height),
    hash          TEXT      NOT NULL
);
CREATE INDEX nyx_nm_v2_ms_status_height_index ON nyx_nym_mixnet_v2_mixnode_status (height);
CREATE INDEX nyx_nm_v2_ms_executed_at_index ON nyx_nym_mixnet_v2_mixnode_status (executed_at);

CREATE TABLE nyx_nym_mixnet_v2_events
(
    -- values: bond, unbond, delegate, undelegate, claim
    event_kind    TEXT      NOT NULL,
    -- values: mixnode_operator, mixnode_delegator, mixnet_rewarding, mixnet_monitoring, gateway_operator
    actor         TEXT      NOT NULL,
    sender        TEXT      NOT NULL,
    proxy         TEXT      NULL,
    mix_id        BIGINT    NULL REFERENCES nyx_nym_mixnet_v2_mixnode (mix_id),
    identity_key  TEXT      NOT NULL,
    executed_at   TIMESTAMP NOT NULL,
    height        BIGINT    NOT NULL REFERENCES block (height),
    hash          TEXT      NOT NULL,
    message_index BIGINT    NOT NULL
);
CREATE INDEX nyx_nm_v2_e_height_index ON nyx_nym_mixnet_v2_events (mix_id, height);
CREATE INDEX nyx_nm_v2_e_executed_at_index ON nyx_nym_mixnet_v2_events (mix_id, executed_at);
CREATE INDEX nyx_nm_v2_e_identity_key_executed_at_index ON nyx_nym_mixnet_v2_events (identity_key, executed_at);

CREATE TABLE nyx_nym_mixnet_v2_mixnode_reward
(
    mix_id            BIGINT    NOT NULL REFERENCES nyx_nym_mixnet_v2_mixnode (mix_id),
    operator_reward   COIN[]    NOT NULL DEFAULT '{}',
    delegates_reward  COIN[]    NOT NULL DEFAULT '{}',
    prior_delegates   COIN[]    NOT NULL DEFAULT '{}',
    prior_unit_reward BIGINT    NOT NULL,
    apy               FLOAT     NOT NULL,
    epoch             BIGINT    NOT NULL,
    executed_at       TIMESTAMP NOT NULL,
    height            BIGINT    NOT NULL REFERENCES block (height),
    hash              TEXT      NOT NULL,
    message_index     BIGINT    NOT NULL,
    UNIQUE (mix_id, height, hash, message_index)
);
CREATE INDEX nyx_nm_v2_mr_height_index ON nyx_nym_mixnet_v2_mixnode_reward (mix_id, height);
CREATE INDEX nyx_nm_v2_mr_executed_at_index ON nyx_nym_mixnet_v2_mixnode_reward (mix_id, executed_at);