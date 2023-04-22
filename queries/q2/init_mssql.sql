if db_id('q2') is not null
begin
    create database q2;
end
go

use test;
go

if object_id('dbo.q2', 'u') is not null drop table dbo.q2;    
go

create table dbo.q2 (
    id int not null,
    data char(100) not null,
    status char(7) not null
);
go

with id as (
    select row_number() over (order by (select 1)) as id
    from sys.all_columns as t1
    cross join sys.all_columns as t2
)
insert into dbo.q2 (id, data, status)
select top (10000000)
    id.id,
    replicate('a', 100),
    case when id.id % 10 = 0 then 'deleted' else 'active' end as status
from id;
go
