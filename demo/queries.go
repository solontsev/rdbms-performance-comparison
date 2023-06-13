package main

import (
	"context"
	"database/sql"
	"log"
	"time"
)

const MySql = "mysql"
const PostgreSql = "postgres"
const MsSql19 = "mssql-19"
const MsSql22 = "mssql-22"

type testData struct {
	testName  string
	queries   map[string]map[string]string
	f         func(context.Context, *sql.DB, string)
	execCount int
}

var Tests = map[string]testData{
	"01": {
		"lookup by primary key",
		map[string]map[string]string{
			MySql: {
				"first key": "select id from client as c where id = 0;",
				//"middle key":                        "select id from client as c where id = 5000;",
				"last key": "select id from client as c where id = 9999;",
				//"not existing key at the beginning": "select id from client as c where id = 100000;",
				//"not existing key at the end":       "select id from client as c where id = 100000;",
				"lookup_and_agg": "select count(*) from order_detail as od where order_id = 1;",
			},
			PostgreSql: {
				"first key": "select id from client as c where id = 0;",
				//"middle key":                        "select id from client as c where id = 5000;",
				"last key": "select id from client as c where id = 9999;",
				//"not existing key at the beginning": "select id from client as c where id = 100000;",
				//"not existing key at the end":       "select id from client as c where id = 100000;",
				"lookup_and_agg": "select count(*) from order_detail as od where order_id = 1;",
			},
			MsSql19: {
				"first key": "select id from client as c where id = 0;",
				//"middle key":                        "select id from client as c where id = 5000;",
				"last key": "select id from client as c where id = 9999;",
				//"not existing key at the beginning": "select id from client as c where id = 100000;",
				//"not existing key at the end":       "select id from client as c where id = 100000;",
				"lookup_and_agg": "select count(*) from order_detail as od where order_id = 1;",
			},
			MsSql22: {
				"first key": "select id from client as c where id = 0;",
				//"middle key":                        "select id from client as c where id = 5000;",
				"last key": "select id from client as c where id = 9999;",
				//"not existing key at the beginning": "select id from client as c where id = 100000;",
				//"not existing key at the end":       "select id from client as c where id = 100000;",
				"lookup_and_agg": "select count(*) from order_detail as od where order_id = 1;",
			},
		},
		QueryInt,
		3000,
	},
	"02": {
		"lookup by primary key + column not in index",
		map[string]map[string]string{
			MySql: {
				"": "select id, name from client as c where id = 1;",
			},
			PostgreSql: {
				"": "select id, name from client as c where id = 1;",
			},
			MsSql19: {
				"": "select id, name from client as c where id = 1;",
			},
			MsSql22: {
				"": "select id, name from client as c where id = 1;",
			},
		},
		QueryIntAndString,
		3000,
	},
	"03": {
		"min and max",
		map[string]map[string]string{
			MySql: {
				"min":     "select min(id) from client as c;",
				"max":     "select min(id) from client as c;",
				"min-max": "select min(id) + max(id) from client as c;",
			},
			PostgreSql: {
				"min":     "select min(id) from client as c;",
				"max":     "select min(id) from client as c;",
				"min-max": "select min(id) + max(id) from client as c;",
			},
			MsSql19: {
				"min":     "select min(id) from client as c;",
				"max":     "select min(id) from client as c;",
				"min-max": "select min(id) + max(id) from client as c;",
			},
			MsSql22: {
				"min":     "select min(id) from client as c;",
				"max":     "select min(id) from client as c;",
				"min-max": "select min(id) + max(id) from client as c;",
			},
		},
		QueryInt,
		3000,
	},
	"04": {
		"index seek with complex condition",
		map[string]map[string]string{
			MySql: {
				"":                  "select count(*) from client where id >= 1 and id < 10000 and id < 2;",
				"bigger range":      "select count(*) from order_detail where order_id >= 1 and order_id < 10000 and order_id < 2;",
				"much bigger range": "select count(*) from order_detail where order_id >= 1 and order_id < 100000 and order_id < 2;",
				"fixed":             "select count(*) from order_detail where order_id >= 1 and order_id < 2 and order_id < 100000;",
			},
			PostgreSql: {
				"":                  "select count(*) from client where id >= 1 and id < 10000 and id < 2;",
				"bigger range":      "select count(*) from order_detail where order_id >= 1 and order_id < 10000 and order_id < 2;",
				"much bigger range": "select count(*) from order_detail where order_id >= 1 and order_id < 100000 and order_id < 2;",
				"fixed":             "select count(*) from order_detail where order_id >= 1 and order_id < 2 and order_id < 100000;",
			},
			MsSql19: {
				"":                  "select count(*) from client where id >= 1 and id < 10000 and id < 2;",
				"bigger range":      "select count(*) from order_detail where order_id >= 1 and order_id < 10000 and order_id < 2;",
				"much bigger range": "select count(*) from order_detail where order_id >= 1 and order_id < 100000 and order_id < 2;",
				"fixed":             "select count(*) from order_detail where order_id >= 1 and order_id < 2 and order_id < 100000;",
			},
			MsSql22: {
				"":                  "select count(*) from client where id >= 1 and id < 10000 and id < 2;",
				"bigger range":      "select count(*) from order_detail where order_id >= 1 and order_id < 10000 and order_id < 2;",
				"much bigger range": "select count(*) from order_detail where order_id >= 1 and order_id < 100000 and order_id < 2;",
				"fixed":             "select count(*) from order_detail where order_id >= 1 and order_id < 2 and order_id < 100000;",
			},
		},
		QueryInt,
		200,
	},
	"05": {
		"nonclustered index seek vs. scan",
		map[string]map[string]string{
			MySql: {
				"":   "select min(name) from client where country = 'FR';",
				"CY": "select min(name) from client where country = 'CY';",
				"US": "select min(name) from client where country = 'US';",
				//">=US": "select min(name) from client where country >= 'US';",
			},
			PostgreSql: {
				"":   "select min(name) from client where country = 'FR';",
				"CY": "select min(name) from client where country = 'CY';",
				"US": "select min(name) from client where country = 'US';",
			},
			MsSql19: {
				"":   "select min(name) from client where country = 'FR';",
				"CY": "select min(name) from client where country = 'CY';",
				"US": "select min(name) from client where country = 'US';",
				//"forceseek":    "select min(name) from client with (forceseek) where country = 'FR';",
				//"forceseek-CY": "select min(name) from client with (forceseek) where country = 'CY';",
				//"forceseek-US": "select min(name) from client with (forceseek) where country = 'US';",
			},
			MsSql22: {
				"":   "select min(name) from client where country = 'FR';",
				"CY": "select min(name) from client where country = 'CY';",
				"US": "select min(name) from client where country = 'US';",
				//"forceseek":    "select min(name) from client with (forceseek) where country = 'FR';",
				//"forceseek-CY": "select min(name) from client with (forceseek) where country = 'CY';",
				//"forceseek-US": "select min(name) from client with (forceseek) where country = 'US';",
			},
		},
		QueryString,
		200,
	},
	//{
	//	"05",
	//	map[string]map[string]string{
	//		"mysql": {
	//			"": "select count(*) from `order` as o inner join `order_detail` as od on od.order_id = o.id;",
	//		},
	//		"postgres": {
	//			"":             "select count(*) from \"order\" as o inner join order_detail as od on od.order_id = o.id;",
	//			"hashjoin=off": "SET enable_hashjoin = off; select count(*) from \"order\" as o inner join order_detail as od on od.order_id = o.id;",
	//		},
	//		"mssql2022": {
	//			"":          "select count(*) from [order] as o inner join order_detail as od on od.order_id = o.id;",
	//			"loop join": "select count(*) from [order] as o inner loop join order_detail as od on od.order_id = o.id;",
	//		},
	//	},
	//	Query05,
	//	0,
	//},
}

func QueryInt(ctx context.Context, db *sql.DB, query string) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// execute query with context and handle no rows error

	var i int
	if err := db.QueryRowContext(ctx, query).Scan(&i); err != nil {
		if err == sql.ErrNoRows {
			i = -1
		} else {
			log.Fatalf("unable to execute query: %v", err)
		}
	}
	//log.Println("result = ", result)
}

func QueryString(ctx context.Context, db *sql.DB, query string) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// execute query with context and handle no rows error

	var s string
	if err := db.QueryRowContext(ctx, query).Scan(&s); err != nil {
		if err == sql.ErrNoRows {
			s = "N/A"
		} else {
			log.Fatalf("unable to execute query: %v", err)
		}
	}
	//log.Println("result = ", result)
}

func QueryIntAndString(ctx context.Context, db *sql.DB, query string) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var i int
	var s string
	if err := db.QueryRowContext(ctx, query).Scan(&i, &s); err != nil {
		if err == sql.ErrNoRows {
			i = -1
			s = "N/A"
		} else {
			log.Fatalf("unable to execute query: %v", err)
		}
	}
	//log.Println("result = ", result)
}
