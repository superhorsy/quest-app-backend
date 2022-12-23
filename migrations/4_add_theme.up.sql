alter table quests
    add theme varchar(10);

comment on column quests.theme is 'Theme in UI for this quest, null by default';
