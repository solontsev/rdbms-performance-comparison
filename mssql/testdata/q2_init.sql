use test;
go

if object_id('dbo.test_table', 'u') is not null drop table dbo.test_table;    
go

create table dbo.test_table (
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
insert into dbo.test_table (id, data, status)
select top (10000000)
    id.id,
    replicate('a', 100),
    case when id.id % 10 = 0 then 'deleted' else 'active' end as status
from id;
go
