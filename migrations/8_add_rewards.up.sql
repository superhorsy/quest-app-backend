alter table quests
    add rewards VARCHAR;

comment on column quests.rewards is 'Rewards that user gets after quest completion';