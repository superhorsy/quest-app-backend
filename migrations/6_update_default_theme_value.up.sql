alter table quests
    alter column theme set not null;

alter table quests
    alter column theme set default 'standart';