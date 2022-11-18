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
    id            uuid                     DEFAULT uuid_generate_v4(),
    sort          int                      NOT NULL,
    "description" VARCHAR(255)             NOT NULL CHECK ("description" <> ''),
    question      VARCHAR(255)             NOT NULL,
    question_type VARCHAR                  NOT NULL,
    answer_type   VARCHAR                  NOT NULL,
    answer        jsonb                    NOT NULL CHECK (jsonb_typeof(answer) = 'array'),
    created_at    TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at    TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at    TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    PRIMARY KEY (id),
    CONSTRAINT steps_id_unique UNIQUE (id)
);

CREATE TABLE IF NOT EXISTS quest_to_step
(
    quest_id uuid NOT NULL,
    step_id  uuid NOT NULL,
    PRIMARY KEY (quest_id, step_id),
    CONSTRAINT quest_id_step_id_unique UNIQUE (quest_id, step_id),
    CONSTRAINT quest_id_fk_step_id FOREIGN KEY (quest_id) REFERENCES steps (id),
    CONSTRAINT step_id_fk_quest_id FOREIGN KEY (step_id) REFERENCES quests (id)
);

CREATE TABLE IF NOT EXISTS quest_to_email
(
    quest_id uuid  NOT NULL,
    email    email NOT NULL,
    PRIMARY KEY (quest_id, email),
    CONSTRAINT quest_id_email_unique UNIQUE (quest_id, email)
);


