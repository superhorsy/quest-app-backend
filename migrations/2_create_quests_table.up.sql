CREATE TABLE IF NOT EXISTS quests
(
    id          uuid                     DEFAULT uuid_generate_v4(),
    "owner"     uuid                     NOT NULL,
    "name"      VARCHAR(255)             NOT NULL CHECK ("name" <> ''),
    description VARCHAR(255),
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at  TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    PRIMARY KEY (id),
    CONSTRAINT quests_id_unique UNIQUE (id),
    CONSTRAINT owner_fk_users_id FOREIGN KEY (owner) REFERENCES users (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS steps
(
    id               uuid                     DEFAULT uuid_generate_v4(),
    quest_id         uuid                     NOT NULL,
    sort             int                      NOT NULL,
    description      VARCHAR(255)             NOT NULL CHECK (description <> ''),
    question_type    VARCHAR                  NOT NULL,
    question_content VARCHAR(255)             NOT NULL,
    answer_type      VARCHAR                  NOT NULL,
    answer_content   jsonb                    NOT NULL,
    created_at       TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at       TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at       TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    PRIMARY KEY (id),
    CONSTRAINT steps_id_unique UNIQUE (id),
    CONSTRAINT quest_id_fk_quests_id FOREIGN KEY (quest_id) REFERENCES quests (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS quest_to_email
(
    quest_id uuid    NOT NULL,
    email    VARCHAR   NOT NULL,
    "name"     VARCHAR NOT NULL,
    PRIMARY KEY (quest_id, email),
    CONSTRAINT quest_id_email_unique UNIQUE (quest_id, email),
--     restrict quest deletion when it is send to someone
    CONSTRAINT quest_id_fk_quests_id FOREIGN KEY (quest_id) REFERENCES quests (id) ON DELETE RESTRICT
);


