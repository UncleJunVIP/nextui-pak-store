create table installed_paks
(
    name text not null,
    type text not null,
    version  text not null,
    unique (name)
);
