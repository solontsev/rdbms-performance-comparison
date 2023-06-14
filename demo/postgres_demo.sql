-- 01 - lookup by primary key
explain analyze select id from client where id = 1;
explain analyze select id from client where id = 100000;

explain analyze select count(*) from order_detail as od where order_id = 1;

-- 02 - lookup by primary key + column not in index
explain analyze select id, name from client where id = 1;
explain analyze select id, name from client where id = 100000;

-- 03 - min and max
explain analyze select min(id) from client;
explain analyze select max(id) from client;
explain analyze select min(id) + max(id) from client;

-- 04 - index seek with complex condition
explain analyze select count(*) from client where id >= 1 and id < 10000 and id > 9990;
explain analyze select count(*) from order_detail where order_id >= 1 and order_id < 10000 and order_id < 2;

-- prepare stmt(int, int, int) as select count(order_id), '' from order_detail where order_id >= $1 and order_id < $2 and order_id < $3;
-- explain analyze execute stmt(1, 10000, 2);
-- explain analyze execute stmt(1, 1000000, 1000000);
-- deallocate prepare stmt;


-- 05 - nonclustered index seek vs. scan
explain analyze select count(name) from client where country = 'UK'; -- 1, index scan
explain analyze select count(name) from client where country = 'NL'; -- 9, bitmap index scan
explain analyze select count(name) from client where country = 'FR'; -- 90
explain analyze select count(name) from client where country = 'CY'; -- 900
explain analyze select count(name) from client where country = 'US'; -- 4000
explain analyze select count(name) from client where country >= 'US'; -- 7333, seq scan


-- 06 - join 2 sorted tables
explain analyze select count(*) from client as c inner join client_ex as c_ex on c_ex.id = c.id;
explain analyze select count(*) from "order" as o inner join order_detail as od on od.order_id = o.id;

/*
set enable_hashjoin = off;
set enable_hashjoin = on;
set max_parallel_workers_per_gather = 0;

show max_parallel_workers_per_gather;
 */

explain analyze select count(*) from "order" as o inner join (select order_id from order_detail group by order_id) as od on od.order_id = o.id;

explain analyze select count(*) from client as a inner join client as b on a.name < b.name;

------------------------------------------------------------------------------------------------------------------------
explain analyze select order_id from order_detail group by order_id;


-- 07 - grouping
explain analyze select min(min_product_id) from (select order_id, min(product_id) as min_product_id from order_detail group by order_id) as t;

explain analyze select min(min_c2) from (select c1, min(c2) as min_c2 from large_group_by_table group by c1) as t;
