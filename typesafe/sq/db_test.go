package sq_test

import (
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
	. "github.com/smartystreets/goconvey/convey"

	core "github.com/ng-vu/sqlgen/core"
	"github.com/ng-vu/sqlgen/mock"
	. "github.com/ng-vu/sqlgen/typesafe/sq"
)

var (
	db   *Database
	merr = new(mock.ErrorMock)
)

func init() {
	connStr := "port=15432 user=sqlgen password=sqlgen dbname=sqlgen sslmode=disable connect_timeout=10"
	db = MustConnect("postgres", connStr, SetErrorMapper(merr.Mock))
	db.MustExec("SELECT 1")
}

func TestDatabase(t *testing.T) {
	Convey("Query", t, func() {
		rows, err := db.Query("SELECT $1::INT + $2::INT", 1, 2)
		So(err, ShouldBeNil)

		var v int
		So(rows.Next(), ShouldBeTrue)
		err = rows.Scan(&v)
		So(err, ShouldBeNil)
		So(v, ShouldEqual, 3)
		So(rows.Next(), ShouldBeFalse)
	})
}

func TestQuery(t *testing.T) {
	Convey("SQL", t, func() {
		Convey("Build", func() {
			query, args, err := db.SQL(`SELECT COUNT(*) FROM "user"`).
				Where("status = ?", "active").Build()
			So(err, ShouldBeNil)
			So(len(args), ShouldEqual, 1)
			So(args[0], ShouldEqual, "active")

			expectedQuery := `SELECT COUNT(*) FROM "user" WHERE (status = $1)`
			So(query, ShouldEqual, expectedQuery)
		})
		Convey("Scan", func() {
			var n int
			err := db.SQL(`SELECT 1`).Where("1 = ?", 1).Scan(&n)
			So(err, ShouldBeNil)
			So(n, ShouldEqual, 1)
		})
	})
	Convey("In", t, func() {
		query, args, err := db.SQL(`SELECT * FROM "user"`).
			Where("status = ?", "active").
			In("id", []int64{10, 20, 30}).
			In(`a."order"`, []string{"40"}).Build()
		So(err, ShouldBeNil)
		So(len(args), ShouldEqual, 5)
		So(args, ShouldResemble, []interface{}{
			"active", int64(10), int64(20), int64(30), "40"})

		expectedQuery := `SELECT * FROM "user" WHERE (status = $1) AND ("id" IN ($2,$3,$4)) AND (a."order" IN ($5))`
		So(query, ShouldEqual, expectedQuery)
	})
	Convey("Tx", t, func() {
		tx, err := db.Begin()
		So(err, ShouldBeNil)
		defer tx.Commit()

		var a, b int
		err = tx.SQL(`SELECT 1`).Scan(&a)
		So(err, ShouldBeNil)

		err = tx.SQL(`SELECT 2`).Scan(&b)
		So(err, ShouldBeNil)

		So(a, ShouldEqual, 1)
		So(b, ShouldEqual, 2)
	})
	Convey("UpdateMap", t, func() {
		query, args, err := db.NewQuery().
			Table("mytable").
			Where("foo = ? AND bar = ?", 1, 2).
			BuildUpdate(core.Map{
				Table: "mytable",
				M: map[string]interface{}{
					"primary": false,
				},
			})
		So(err, ShouldBeNil)
		So(args, ShouldResemble, []interface{}{false, 1, 2})

		expectedQuery := `UPDATE "mytable" SET "primary" = $1 WHERE (foo = $2 AND bar = $3)`
		So(query, ShouldEqual, expectedQuery)
	})
}

func TestErrorMapper(t *testing.T) {
	merr.Reset()
	Convey("ErrorMapper", t, func() {
		Reset(func() {
			merr.Reset()
		})
		Convey("Query", func() {
			Convey("Exec", func() {
				_, err := db.Exec("SELECT a")
				_, ok := err.(*mock.Error)
				So(ok, ShouldBeTrue)

				So(merr.Called, ShouldEqual, 1)
				So(merr.Err, ShouldBeError, `pq: column "a" does not exist`)
				So(merr.Entry.Query, ShouldEqual, "SELECT a")
				So(merr.Entry.Type(), ShouldEqual, TypeExec)
				So(merr.Entry.IsQuery(), ShouldEqual, true)
				So(merr.Entry.IsTx(), ShouldEqual, false)
			})
			Convey("Query", func() {
				_, err := db.Query("SELECT a")
				_, ok := err.(*mock.Error)
				So(ok, ShouldBeTrue)

				So(merr.Called, ShouldEqual, 1)
				So(merr.Err, ShouldBeError, `pq: column "a" does not exist`)
				So(merr.Entry.Query, ShouldEqual, "SELECT a")
				So(merr.Entry.Type(), ShouldEqual, TypeQuery)
				So(merr.Entry.IsQuery(), ShouldEqual, true)
				So(merr.Entry.IsTx(), ShouldEqual, false)
			})
			Convey("QueryRow", func() {
				err := db.QueryRow("SELECT a").Scan()
				_, ok := err.(*mock.Error)
				So(ok, ShouldBeTrue)

				So(merr.Called, ShouldEqual, 1)
				So(merr.Err, ShouldBeError, `pq: column "a" does not exist`)
				So(merr.Entry.Query, ShouldEqual, "SELECT a")
				So(merr.Entry.Type(), ShouldEqual, TypeQueryRow)
				So(merr.Entry.IsQuery(), ShouldEqual, true)
				So(merr.Entry.IsTx(), ShouldEqual, false)
			})
			Convey("QueryRow - sql.NoRows", func() {
				err := db.QueryRow("SELECT 1 WHERE false").Scan()
				_, ok := err.(*mock.Error)
				So(ok, ShouldBeTrue)

				So(merr.Called, ShouldEqual, 1)
				So(merr.Err, ShouldEqual, sql.ErrNoRows)
				So(merr.Entry.Query, ShouldEqual, "SELECT 1 WHERE false")
				So(merr.Entry.Type(), ShouldEqual, TypeQueryRow)
				So(merr.Entry.IsQuery(), ShouldEqual, true)
				So(merr.Entry.IsTx(), ShouldEqual, false)
			})
		})
		Convey("Tx", func() {
			tx, err := db.Begin()
			So(err, ShouldBeNil)
			Reset(func() {
				tx.Rollback()
			})
			Convey("Exec", func() {
				_, err := tx.Exec("SELECT a")
				_, ok := err.(*mock.Error)
				So(ok, ShouldBeTrue)

				So(merr.Called, ShouldEqual, 1)
				So(merr.Err, ShouldBeError, `pq: column "a" does not exist`)
				So(merr.Entry.Query, ShouldEqual, "SELECT a")
				So(merr.Entry.Type(), ShouldEqual, TypeExec)
				So(merr.Entry.IsQuery(), ShouldEqual, true)
				So(merr.Entry.IsTx(), ShouldEqual, true)
			})
			Convey("Query", func() {
				_, err := tx.Query("SELECT a")
				_, ok := err.(*mock.Error)
				So(ok, ShouldBeTrue)

				So(merr.Called, ShouldEqual, 1)
				So(merr.Err, ShouldBeError, `pq: column "a" does not exist`)
				So(merr.Entry.Query, ShouldEqual, "SELECT a")
				So(merr.Entry.Type(), ShouldEqual, TypeQuery)
				So(merr.Entry.IsQuery(), ShouldEqual, true)
				So(merr.Entry.IsTx(), ShouldEqual, true)
			})
			Convey("QueryRow", func() {
				err := tx.QueryRow("SELECT a").Scan()
				_, ok := err.(*mock.Error)
				So(ok, ShouldBeTrue)

				So(merr.Called, ShouldEqual, 1)
				So(merr.Err, ShouldBeError, `pq: column "a" does not exist`)
				So(merr.Entry.Query, ShouldEqual, "SELECT a")
				So(merr.Entry.Type(), ShouldEqual, TypeQueryRow)
				So(merr.Entry.IsQuery(), ShouldEqual, true)
				So(merr.Entry.IsTx(), ShouldEqual, true)
			})
			Convey("QueryRow - sql.NoRows", func() {
				err := tx.QueryRow("SELECT 1 WHERE false").Scan()
				_, ok := err.(*mock.Error)
				So(ok, ShouldBeTrue)

				So(merr.Called, ShouldEqual, 1)
				So(merr.Err, ShouldEqual, sql.ErrNoRows)
				So(merr.Entry.Query, ShouldEqual, "SELECT 1 WHERE false")
				So(merr.Entry.Type(), ShouldEqual, TypeQueryRow)
				So(merr.Entry.IsQuery(), ShouldEqual, true)
				So(merr.Entry.IsTx(), ShouldEqual, true)
			})
			Convey("Commit", func() {
				tx.Exec("SELECT a")
				tx.Exec("SELECT b")

				err := tx.Commit()
				So(err, ShouldNotBeNil)

				So(merr.Called, ShouldEqual, 3)
				So(merr.Err, ShouldNotBeNil)
				So(merr.Entry.Type(), ShouldEqual, TypeCommit)
				So(merr.Entry.IsQuery(), ShouldEqual, false)
				So(merr.Entry.IsTx(), ShouldEqual, true)
				So(len(merr.Entry.TxQueries), ShouldEqual, 2)
				So(merr.Entry.TxQueries[0].Query, ShouldEqual, "SELECT a")
				So(merr.Entry.TxQueries[1].Query, ShouldEqual, "SELECT b")
			})
			Convey("Rollback", func() {
				tx.Exec("SELECT a")
				tx.Exec("SELECT b")

				err := tx.Rollback()
				So(err, ShouldBeNil)

				So(merr.Called, ShouldEqual, 3)
				So(merr.Err, ShouldBeNil)
				So(merr.Entry.Type(), ShouldEqual, TypeRollback)
				So(merr.Entry.IsQuery(), ShouldEqual, false)
				So(merr.Entry.IsTx(), ShouldEqual, true)
				So(len(merr.Entry.TxQueries), ShouldEqual, 2)
				So(merr.Entry.TxQueries[0].Query, ShouldEqual, "SELECT a")
				So(merr.Entry.TxQueries[1].Query, ShouldEqual, "SELECT b")
			})
		})
		Convey("Build", func() {
			_, err := db.Table("foo").UpdateMap(nil)
			So(err, ShouldBeError, "sqlgen: UPDATE must have WHERE")

			So(merr.Called, ShouldEqual, 1)
			So(merr.Entry.IsBuild(), ShouldEqual, true)
			So(merr.Entry.IsQuery(), ShouldEqual, true)
			So(merr.Entry.IsTx(), ShouldEqual, false)
		})
		Convey("Build - Tx", func() {
			tx, err := db.Begin()
			So(err, ShouldBeNil)

			_, err = tx.Table("foo").UpdateMap(nil)
			So(err, ShouldBeError, "sqlgen: UPDATE must have WHERE")

			So(merr.Called, ShouldEqual, 1)
			So(merr.Entry.IsBuild(), ShouldEqual, true)
			So(merr.Entry.IsQuery(), ShouldEqual, true)
			So(merr.Entry.IsTx(), ShouldEqual, true)
		})
	})
}
