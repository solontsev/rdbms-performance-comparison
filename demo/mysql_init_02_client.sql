use test_db;

drop table if exists `client`;
drop table if exists `client_ex`;

create table `client` (
    id int not null,
    name varchar(100) not null,
    country char(2) not null,
    insert_dt datetime not null,

    primary key (id)
);

create table `client_ex` (
    id int not null,
    address varchar(1000) null,

    primary key (id)
);

insert into `client`(id, name, country, insert_dt)
select
    id,
    concat('client_', id),
    case
        when id % 100 = 0 then 'FR'
        when id % 10 = 0 then 'CY'
        when id % 2 = 0 then 'US'
        when id % 3 = 0 then 'DE'
        else 'UK'
        end,
    date_add('2020-01-01', interval id second)
from numbers;

-- select country, count(*) from client group by country;

create index idx_client_country on `client`(country);

insert into `client_ex`(id, address)
select
    id,
    concat('client_', id, '_', repeat('x', 900))
from numbers;