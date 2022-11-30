alter table quest_to_email
    add status VARCHAR default 'not_started' not null;

alter table quest_to_email
    add current_step integer default 0;

-- alter table quests
--     add theme varchar(10);

-- comment on column quests.theme is 'Theme in UI for this quest, null by default';
