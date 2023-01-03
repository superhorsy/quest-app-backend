alter table steps
    add constraint steps_pk
        unique (quest_id, sort);