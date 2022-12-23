alter table quests
    alter column theme drop not null;

alter table quests
    alter column theme drop default;