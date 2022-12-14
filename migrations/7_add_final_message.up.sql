alter table quests
    add final_message varchar;

comment on column quests.final_message is 'Message after quest completion';