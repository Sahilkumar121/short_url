-- 1. drop the foreign key contraint
alter table urls
drop constraint if exists fk_urls_users;

-- 2. drop the column
alter table urls
drop column if exists user_id;
