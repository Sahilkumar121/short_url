alter table urls
add column user_id UUID;

alter table urls
add constraint fk_urls_users
foreign key(user_id)
references users(id)
on delete set null;
