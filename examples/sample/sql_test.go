package test

import (
	"database/sql"
	"encoding/json"
	"log"
	"regexp"
	"testing"
	"time"

	_ "github.com/lib/pq"
	. "github.com/ng-vu/goconveyx"
	. "github.com/smartystreets/goconvey/convey"

	mock "github.com/ng-vu/sqlgen/mock"
	sq "github.com/ng-vu/sqlgen/typesafe/sq"
)

var (
	db   *sq.Database
	merr = new(mock.ErrorMock)

	now0, now1 time.Time
)

type (
	S []interface{}
	M map[string]interface{}
)

func init() {
	connStr := "port=15432 user=sqlgen password=sqlgen dbname=sqlgen sslmode=disable connect_timeout=10"
	db = sq.MustConnect("postgres", connStr, sq.SetErrorMapper(merr.Mock))
	db.MustExec("SELECT 1")

	InitSchema()
	log.Println("Initialized database for testing")

	location, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		log.Fatalf("Unable to load timezone %v", err)
	}
	now0 = time.Date(2020, 10, 11, 8, 9, 10, 123e6, location)
	now1 = time.Date(2021, 10, 11, 8, 9, 10, 123e6, location)
}

func InitSchema() {
	db.MustExec(`
		DROP TABLE IF EXISTS "user_info", "complex_info", "user";
        CREATE TABLE "user" (
            id TEXT PRIMARY KEY,
            name       TEXT,
            created_at TIMESTAMPTZ,
            updated_at TIMESTAMPTZ,
            bool       BOOLEAN,
            float64    DOUBLE PRECISION,
            int        INTEGER,
            int64      BIGINT,
            string     TEXT,
            p_bool     BOOLEAN,
            p_float64  DOUBLE PRECISION,
            p_int      INTEGER,
            p_int64    BIGINT,
            p_string   TEXT
		);
		CREATE TABLE "user_info" (
			user_id    TEXT PRIMARY KEY,
			metadata   TEXT,
			bool       BOOLEAN,
            float64    DOUBLE PRECISION,
            int        INTEGER,
            int64      BIGINT,
            string     TEXT,
            p_bool     BOOLEAN,
            p_float64  DOUBLE PRECISION,
            p_int      INTEGER,
            p_int64    BIGINT,
            p_string   TEXT
		);
		CREATE TABLE "complex_info" (
			id TEXT PRIMARY KEY,
			address         JSONB,
			p_address       JSONB,
			metadata        JSONB,
			ints            INTEGER[],
			int64s          BIGINT[],
			strings         TEXT[],
			times           TIMESTAMPTZ[],
			times_p         TIMESTAMPTZ[],
			alias_string    TEXT,
			alias_int64     BIGINT,
			alias_int       INTEGER,
			alias_bool      BOOLEAN,
			alias_float64   DOUBLE PRECISION,
			-- alias_time      TIMESTAMPTZ,
			alias_p_string  TEXT,
			alias_p_int64   BIGINT,
			alias_p_int     INTEGER,
			alias_p_bool    BOOLEAN,
			alias_p_float64 DOUBLE PRECISION
			-- alias_p_time    TIMESTAMPTZ
		);
	`)
}

func TestUser(t *testing.T) {
	Convey("Insert", t, func() {
		Reset(func() {
			db.MustExec(`TRUNCATE "user"`)
		})

		users := []*User{
			{
				ID:        "1000",
				Name:      "Alice",
				CreatedAt: now0,
				UpdatedAt: &now1,
				Bool:      true,
				Float64:   1000.1,
				Int:       1001,
				Int64:     1002,
				String:    "string",
				PBool:     pBool(false),
				PFloat64:  pFloat64(0),
				PInt:      pInt(0),
				PInt64:    pInt64(0),
				PString:   pString(""),
			}, {
				ID:   "1001",
				Name: "Kattie",
			},
		}
		expectedUsers := []interface{}{
			map[string]interface{}{
				"id":   "1000",
				"name": "Alice",

				// time should be retrieved as UTC
				"created_at": toTime("2020-10-11T01:09:10.123Z"),
				"updated_at": toTime("2021-10-11T01:09:10.123Z"),

				// basic type should be stored correctly
				"bool":    true,
				"float64": float64(1000.1),
				"int":     int64(1001),
				"int64":   int64(1002),
				"string":  "string",

				// non-nil pointer should be stored as value
				"p_bool":    false,
				"p_float64": float64(0),
				"p_int":     int64(0),
				"p_int64":   int64(0),
				"p_string":  "",
			},
			map[string]interface{}{
				"id":   "1001",
				"name": "Kattie",

				// zero time should be stored as null
				"created_at": nil,
				"updated_at": nil,

				// bool, float and int should be stored as non-null
				"bool":    false,
				"float64": float64(0),
				"int":     int64(0),

				// zero int64 and string should be stored as null
				"int64":  nil,
				"string": nil,

				// nil pointer should be stored as null
				"p_bool":    nil,
				"p_float64": nil,
				"p_int":     nil,
				"p_int64":   nil,
				"p_string":  nil,
			},
		}
		{
			n, err := db.Insert(Users(users))
			So(err, ShouldBeNil)
			So(n, ShouldEqual, 2)

			users[0].CreatedAt = users[0].CreatedAt.In(time.UTC)
			*users[0].UpdatedAt = users[0].UpdatedAt.In(time.UTC)
		}
		Convey("Insert: Build", func() {
			query, args, err := db.NewQuery().BuildInsert(Users(users))
			So(err, ShouldBeNil)
			So(len(args), ShouldEqual, 28)

			expectedQuery := `INSERT INTO "user" ("id","name","created_at","updated_at","bool","float64","int","int64","string","p_bool","p_float64","p_int","p_int64","p_string") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14),($15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28)`
			So(query, ShouldEqual, expectedQuery)
		})
		Convey("Get again", func() {
			actual := shouldQuery(db, `SELECT * FROM "user"`)
			So(len(actual), ShouldEqual, 2)
			So(actual, ShouldResembleSlice, expectedUsers)
		})
		Convey("Scan rows", func() {
			var _users []*User
			err := db.Find((*Users)(&_users))
			So(err, ShouldBeNil)
			So(len(_users), ShouldEqual, 2)
			So(_users, ShouldResembleSlice, users)
		})
		Convey("Scan single row with simple where condition", func() {
			{
				var user User
				has, err := db.Where("id = ?", "1000").Get(&user)
				So(err, ShouldBeNil)
				So(has, ShouldBeTrue)
				So(&user, ShouldDeepEqual, users[0])
			}
			{
				var user User
				has, err := db.Where("id = ?", "1001").Get(&user)
				So(err, ShouldBeNil)
				So(has, ShouldBeTrue)
				So(&user, ShouldDeepEqual, users[1])
			}
		})
		Convey("Update no column", func() {
			update := &User{}
			_, _, err := db.Where("id = ?", "1000").BuildUpdate(update)
			So(err, ShouldBeError, "common/sql: No column to update")
		})
		Convey("Update", func() {
			update := &User{
				Name: "Alice in wonderland",
				Int:  100,
			}
			Convey("Build", func() {
				query, args, err := db.Where("id = ?", "1000").BuildUpdate(update)
				So(err, ShouldBeNil)

				expectedQuery := `UPDATE "user" SET "name"=$1,"int"=$2 WHERE (id = $3)`
				So(query, ShouldEqual, expectedQuery)
				So(args, ShouldDeepEqual, []interface{}{
					"Alice in wonderland",
					100,
					"1000",
				})
			})
			Convey("Update data", func() {
				n, err := db.Where("id = ?", "1000").Update(update)
				So(err, ShouldBeNil)
				So(n, ShouldEqual, 1)

				Convey("Get again", func() {
					actual := shouldQuery(db, `SELECT * FROM "user"`)
					So(len(actual), ShouldEqual, 2)

					user0 := expectedUsers[0].(map[string]interface{})
					user0["name"] = "Alice in wonderland"
					user0["int"] = int64(100)
					So(actual, ShouldResembleByKey("id"), expectedUsers)
				})
			})
			Convey("Update all: Build", func() {
				query, args, err := db.Where("id = ?", "1000").UpdateAll().BuildUpdate(update)
				So(err, ShouldBeNil)

				expectedQuery := `UPDATE "user" SET ("id","name","created_at","updated_at","bool","float64","int","int64","string","p_bool","p_float64","p_int","p_int64","p_string") = ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) WHERE (id = $15)`
				So(query, ShouldEqual, expectedQuery)
				So(len(args), ShouldEqual, 15)
				So(args[1], ShouldEqual, "Alice in wonderland")
				So(args[6], ShouldEqual, 100)
				So(args[14], ShouldEqual, "1000")
			})
		})
		Convey("Delete", func() {
			q := db.Where("id = ?", "1000")
			Convey("Build", func() {
				query, args, err := q.BuildDelete(&User{})
				So(err, ShouldBeNil)

				expectedQuery := `DELETE FROM "user" WHERE (id = $1)`
				So(query, ShouldEqual, expectedQuery)
				So(args, ShouldDeepEqual, []interface{}{"1000"})
			})
			Convey("Delete data", func() {
				n, err := q.Delete(&User{})
				So(err, ShouldBeNil)
				So(n, ShouldEqual, 1)
				{
					var count int
					err := db.QueryRow(`SELECT COUNT(*) FROM "user"`).Scan(&count)
					So(err, ShouldBeNil)
					So(count, ShouldEqual, 1)
				}
			})
		})
		Convey("Join", func() {
			Convey("Build", func() {
				var userUnion UserUnion
				query, _, err := db.NewQuery().BuildGet(&userUnion)
				So(err, ShouldBeNil)

				expectedQuery := `SELECT u."id",u."name",u."created_at",u."updated_at",u."bool",u."float64",u."int",u."int64",u."string",u."p_bool",u."p_float64",u."p_int",u."p_int64",u."p_string",ui."user_id",ui."metadata",ui."bool",ui."float64",ui."int",ui."int64",ui."string",ui."p_bool",ui."p_float64",ui."p_int",ui."p_int64",ui."p_string" FROM "user" AS u FULL OUTER JOIN "user_info" AS ui ON u.id = ui.user_id`
				So(query, ShouldEqual, expectedQuery)
			})

			Reset(func() {
				db.MustExec(`TRUNCATE "user_info"`)
			})
			Convey("Scan", func() {
				var userUnion UserUnion
				has, err := db.Where(`u.id = ?`, "1000").Get(&userUnion)
				So(err, ShouldBeNil)
				So(has, ShouldEqual, true)
				So(userUnion.User, ShouldDeepEqual, users[0])
			})
			Convey("Scan rows", func() {
				var userUnions UserUnions
				err := db.Find(&userUnions)
				So(err, ShouldBeNil)
				So(len(userUnions), ShouldEqual, 2)

				_users := []*User{userUnions[0].User, userUnions[1].User}
				So(_users, ShouldResembleByKey("ID"), users)
			})
		})
	})
	Convey("Array & JSON", t, func() {
		Reset(func() {
			db.MustExec(`TRUNCATE "complex_info"`)
		})

		address := Address{
			Province: "p",
		}
		items := []*ComplexInfo{
			{
				ID:       "1000",
				Address:  address,
				PAddress: &address,
				Metadata: map[string]string{
					"foo": "bar",
				},
				Ints:    []int{1, 2, 3},
				Int64s:  []int64{4, 5, 6},
				Strings: []string{"a", "b", "c"},
				Times:   []time.Time{now0, now1},
				TimesP:  []*time.Time{&now0, &now1},

				AliasBool:    AliasBool(true),
				AliasFloat64: AliasFloat64(1.23),
				AliasInt:     AliasInt(777),
				AliasInt64:   AliasInt64(int64(999999)),
				AliasString:  AliasString("fake"),
				// AliasTime:    AliasTime(now0),

				AliasPBool:    AliasPBool(pBool(false)),
				AliasPFloat64: AliasPFloat64(pFloat64(3.45)),
				AliasPInt:     AliasPInt(pInt(333)),
				AliasPInt64:   AliasPInt64(pInt64(88888888888)),
				AliasPString:  AliasPString(pString("pfake")),
				// AliasPTime:    AliasPTime(&now1),
			},
			{
				ID: "1001",
			},
		}

		expectedTimes := `{"2020-10-11 01:09:10.123+00","2021-10-11 01:09:10.123+00"}`
		expectedItems := []interface{}{
			map[string]interface{}{
				"id":        "1000",
				"address":   []byte(`{"province": "p"}`),
				"p_address": []byte(`{"province": "p"}`),
				"metadata":  []byte(`{"foo": "bar"}`),

				"ints":    []byte("{1,2,3}"),
				"int64s":  []byte("{4,5,6}"),
				"strings": []byte("{a,b,c}"),
				"times":   []byte(expectedTimes),
				"times_p": []byte(expectedTimes),

				"alias_bool":    true,
				"alias_float64": float64(1.23),
				"alias_int":     int64(777),
				"alias_int64":   int64(999999),
				"alias_string":  "fake",
				// "alias_time":    toTime("2020-10-11T01:09:10.123Z"),

				"alias_p_bool":    false,
				"alias_p_float64": float64(3.45),
				"alias_p_int":     int64(333),
				"alias_p_int64":   int64(88888888888),
				"alias_p_string":  "pfake",
				// "alias_p_time":    toTime("2021-10-11T01:09:10.123Z"),
			},
			map[string]interface{}{
				"id":        "1001",
				"address":   []byte(`{"province": ""}`),
				"p_address": nil,
				"metadata":  nil,

				"ints":    nil,
				"int64s":  nil,
				"strings": nil,
				"times":   nil,
				"times_p": nil,

				"alias_bool":    false,
				"alias_float64": float64(0),
				"alias_int":     int64(0),

				// zero int64 and string should be stored as null
				"alias_int64":  nil,
				"alias_string": nil,
				// "alias_time":   nil,

				"alias_p_bool":    nil,
				"alias_p_float64": nil,
				"alias_p_int":     nil,
				"alias_p_int64":   nil,
				"alias_p_string":  nil,
				// "alias_p_time":    nil,
			},
		}
		{
			n, err := db.Insert(ComplexInfoes(items))
			So(err, ShouldBeNil)
			So(n, ShouldEqual, 2)

			items[0].Times[0] = now0.In(time.UTC)
			items[0].Times[1] = now1.In(time.UTC)
			items[0].TimesP[0] = pTime(now0.In(time.UTC))
			items[0].TimesP[1] = pTime(now1.In(time.UTC))
			// items[0].AliasTime = AliasTime(now0.In(time.UTC))
			// items[0].AliasPTime = AliasPTime(pTime(now1.In(time.UTC)))
		}
		Convey("Insert: Build", func() {
			query, args, err := db.NewQuery().BuildInsert(ComplexInfoes(items))
			So(err, ShouldBeNil)
			So(len(args), ShouldEqual, 38)

			expectedQuery := `INSERT INTO "complex_info" ("id","address","p_address","metadata","ints","int64s","strings","times","times_p","alias_string","alias_int64","alias_int","alias_bool","alias_float64","alias_p_string","alias_p_int64","alias_p_int","alias_p_bool","alias_p_float64") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19),($20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33,$34,$35,$36,$37,$38)`
			So(query, ShouldEqual, expectedQuery)
		})
		Convey("Get again", func() {
			actual := shouldQuery(db, `SELECT * FROM "complex_info"`)
			So(len(actual), ShouldEqual, 2)
			So(actual, ShouldResembleByKey("id"), expectedItems)
		})
		Convey("Scan", func() {
			actual := new(ComplexInfo)
			has, err := db.Where("id = ?", "1000").Get(actual)
			So(err, ShouldBeNil)
			So(has, ShouldBeTrue)
			So(actual, ShouldDeepEqual, items[0])
		})
		Convey("Scan rows", func() {
			var actual ComplexInfoes
			err := db.Find(&actual)
			So(err, ShouldBeNil)
			So(actual, ShouldResembleByKey("ID"), items)
		})
	})
	Convey("Scan null values", t, func() {
		Reset(func() {
			db.MustExec(`TRUNCATE "user", "complex_info"`)
		})
		db.MustExec(`INSERT INTO "user" ("id") VALUES ($1)`, "1000")
		db.MustExec(`INSERT INTO "complex_info" ("id") VALUES ($1)`, "1000")

		{
			var users Users
			err := db.Find(&users)
			So(err, ShouldBeNil)
		}
		{
			var items ComplexInfoes
			err := db.Find(&items)
			So(err, ShouldBeNil)
		}
	})
	Convey("Select and count with JOIN", t, func() {
		query, _, err := db.
			Select("order").
			Where("u.id = ui.user_id").
			BuildCount((*UserUnion)(nil))
		So(err, ShouldBeNil)

		expectedQuery := `SELECT "order",COUNT(*) FROM "user" AS u FULL OUTER JOIN "user_info" AS ui ON u.id = ui.user_id WHERE (u.id = ui.user_id)`
		So(query, ShouldEqual, expectedQuery)
	})
	Convey("Select and count without JOIN", t, func() {
		query, _, err := db.
			Select("order").
			From("user").
			Where("status = ?", "active").
			BuildCount((*User)(nil))
		So(err, ShouldBeNil)

		expectedQuery := `SELECT "order",COUNT(*) FROM "user" WHERE (status = $1)`
		So(query, ShouldEqual, expectedQuery)
	})
}

func TestErrorMapper(t *testing.T) {
	merr.Reset()
	Convey("ErrorMapper", t, func() {
		Reset(func() {
			merr.Reset()
		})

		Convey("Scan: No error", func() {
			var v int
			err := db.Select("1").Scan(&v)
			So(err, ShouldBeNil)

			So(merr.Called, ShouldEqual, 1)
			So(merr.Entry.Query, ShouldEqual, `SELECT 1`)
		})
		Convey("Scan: ErrNoRows", func() {
			var v int
			err := db.Select("1").Where("false").Scan(&v)
			So(err, ShouldNotBeNil)

			So(merr.Called, ShouldEqual, 1)
			So(merr.Entry.Query, ShouldEqual, `SELECT 1 WHERE (false)`)
			So(merr.Err, ShouldEqual, sql.ErrNoRows)
		})
		Convey("Scan: Other error", func() {
			var v int
			err := db.Select("a").Scan(&v)
			So(err, ShouldNotBeNil)

			So(merr.Called, ShouldEqual, 1)
			So(merr.Entry.Query, ShouldEqual, `SELECT "a"`)
			So(merr.Err, ShouldBeError, `pq: column "a" does not exist`)
		})
		Convey("Get: No error", func() {
			user := &User{ID: "foo"}
			{
				_, err := db.Insert(user)
				So(err, ShouldBeNil)
			}

			_, err := db.Get(user)
			So(err, ShouldBeNil)

			So(merr.Called, ShouldEqual, 2)
			So(merr.Entry.Query, ShouldContainSubstring, `FROM "user"`)
		})
		Convey("Get: ErrNoRows", func() {
			var user User
			_, err := db.Where("false").Get(&user)
			So(err, ShouldBeNil)

			So(merr.Called, ShouldEqual, 1)
			So(merr.Entry.Query, ShouldContainSubstring, `FROM "user"`)
			So(merr.Err, ShouldEqual, sql.ErrNoRows)
		})
		Convey("Get: Other error", func() {
			var user User
			_, err := db.Where("invalid").Get(&user)
			So(err, ShouldNotBeNil)

			So(merr.Called, ShouldEqual, 1)
			So(merr.Entry.Query, ShouldContainSubstring, `FROM "user"`)
			So(merr.Err, ShouldBeError, `pq: column "invalid" does not exist`)
		})
	})
}

func pBool(v bool) *bool           { return &v }
func pFloat64(v float64) *float64  { return &v }
func pInt(v int) *int              { return &v }
func pInt64(v int64) *int64        { return &v }
func pString(v string) *string     { return &v }
func pTime(v time.Time) *time.Time { return &v }

var reClean = regexp.MustCompile(`\s+`)

func clean(s string) string {
	return reClean.ReplaceAllString(s, "")
}

func cleans(ss []string) []string {
	for i, s := range ss {
		ss[i] = clean(s)
	}
	return ss
}

func unmmarshal(s string) interface{} {
	var v interface{}
	err := json.Unmarshal([]byte(s), &v)
	if err != nil {
		panic(err)
	}
	return v
}

func jSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func toTime(s string) time.Time {
	var t time.Time
	err := json.Unmarshal([]byte(`"`+s+`"`), &t)
	if err != nil {
		panic(err)
	}
	return t
}

func toPTime(s string) *time.Time {
	t := toTime(s)
	return &t
}

func query(db *sq.Database, sql string, args ...interface{}) ([]interface{}, error) {
	rows, err := db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		rows.Close()
	}()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var res []interface{}
	for rows.Next() {
		row := make([]interface{}, len(cols))
		scan := make([]interface{}, len(cols))
		for i := range cols {
			scan[i] = &row[i]
		}
		err := rows.Scan(scan...)
		if err != nil {
			return nil, err
		}

		m := make(map[string]interface{})
		for i, col := range cols {
			m[col] = row[i]
		}
		res = append(res, m)
	}
	return res, rows.Err()
}

func shouldQuery(db *sq.Database, sql string, args ...interface{}) []interface{} {
	s, err := query(db, sql, args...)
	if err != nil {
		log.Panicf("%v: sql=`%v` args=%#v", err, sql, interface{}(args))
	}
	return s
}
