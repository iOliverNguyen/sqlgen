package sq_test

import (
	"testing"

	"github.com/ng-vu/sqlgen/core"
	. "github.com/ng-vu/sqlgen/typesafe/sq"

	. "github.com/smartystreets/goconvey/convey"
)

type A = []interface{}

func byID(ID int64) *ColumnFilter {
	return &ColumnFilter{
		Prefix: `"user"`,
		Column: "id",
		Value:  ID,
		IsNil:  ID == 0,
	}
}

func byIDPtr(ID *int64) *ColumnFilterPtr {
	return &ColumnFilterPtr{
		Prefix: `"user"`,
		Column: "id",
		Value:  ID,
		IsNil:  ID == nil,
		IsZero: ID != nil && (*ID) == 0,
	}
}

func compose(writerTos ...WriterTo) (string, []interface{}, error) {
	w := NewWriter(core.Opts{}, '"', '$', 1024)
	for _, wt := range writerTos {
		err := wt.WriteSQLTo(w)
		if err != nil {
			return "", nil, err
		}
	}
	return w.String(), w.Args(), nil
}

func TestFilter(t *testing.T) {
	Convey("ByID", t, func() {
		Convey("ByID (default)", func() {
			Convey("empty (should error)", func() {
				_, _, err := compose(byID(0))
				So(err, ShouldBeError, "missing id")
			})
			Convey("not empty", func() {
				sql, args, err := compose(byID(123))
				So(err, ShouldBeNil)
				So(sql, ShouldEqual, `"user"."id" = $1`)
				So(args, ShouldResemble, A{int64(123)})
			})
		})
		Convey("ByID (optional)", func() {
			Convey("empty (should skip)", func() {
				sql, args, err := compose(byID(0).Optional())
				So(err, ShouldBeNil)
				So(sql, ShouldEqual, ``)
				So(args, ShouldResemble, A{})
			})
			Convey("not empty", func() {
				sql, args, err := compose(byID(123).Optional())
				So(err, ShouldBeNil)
				So(sql, ShouldEqual, `"user"."id" = $1`)
				So(args, ShouldResemble, A{int64(123)})
			})
		})
		Convey("ByID (nullable)", func() {
			Convey("empty (should be null)", func() {
				sql, args, err := compose(byID(0).Nullable())
				So(err, ShouldBeNil)
				So(sql, ShouldEqual, `"user"."id" IS NULL`)
				So(args, ShouldResemble, A{})
			})
			Convey("not empty", func() {
				sql, args, err := compose(byID(123).Nullable())
				So(err, ShouldBeNil)
				So(sql, ShouldEqual, `"user"."id" = $1`)
				So(args, ShouldResemble, A{int64(123)})
			})
		})
	})
	Convey("ByIDPtr", t, func() {
		Convey("ByIDPtr (default)", func() {
			Convey("nil (should error)", func() {
				_, _, err := compose(byIDPtr(nil))
				So(err, ShouldBeError, "missing id")
			})
			Convey("zero", func() {
				var id int64 = 0
				sql, args, err := compose(byIDPtr(&id))
				So(err, ShouldBeNil)
				So(sql, ShouldEqual, `"user"."id" IS NULL OR "user"."id" = $1`)
				So(args, ShouldResemble, A{&id})
			})
			Convey("not empty", func() {
				var id int64 = 123
				sql, args, err := compose(byIDPtr(&id))
				So(err, ShouldBeNil)
				So(sql, ShouldEqual, `"user"."id" = $1`)
				So(args, ShouldResemble, A{&id})
			})
		})
		Convey("ByIDPtr (optional)", func() {
			Convey("nil (should skip)", func() {
				sql, args, err := compose(byIDPtr(nil).Optional())
				So(err, ShouldBeNil)
				So(sql, ShouldEqual, ``)
				So(args, ShouldResemble, A{})
			})
			Convey("zero", func() {
				var id int64 = 0
				sql, args, err := compose(byIDPtr(&id).Optional())
				So(err, ShouldBeNil)
				So(sql, ShouldEqual, `"user"."id" IS NULL OR "user"."id" = $1`)
				So(args, ShouldResemble, A{&id})
			})
			Convey("not empty", func() {
				var id int64 = 123
				sql, args, err := compose(byIDPtr(&id).Optional())
				So(err, ShouldBeNil)
				So(sql, ShouldEqual, `"user"."id" = $1`)
				So(args, ShouldResemble, A{&id})
			})
		})
		Convey("ByIDPtr (nullable)", func() {
			Convey("nil (should be null)", func() {
				sql, args, err := compose(byIDPtr(nil).Nullable())
				So(err, ShouldBeNil)
				So(sql, ShouldEqual, `"user"."id" IS NULL`)
				So(args, ShouldResemble, A{})
			})
			Convey("zero", func() {
				var id int64 = 0
				sql, args, err := compose(byIDPtr(&id).Nullable())
				So(err, ShouldBeNil)
				So(sql, ShouldEqual, `"user"."id" = $1`)
				So(args, ShouldResemble, A{&id})
			})
			Convey("not empty", func() {
				var id int64 = 123
				sql, args, err := compose(byIDPtr(&id).Nullable())
				So(err, ShouldBeNil)
				So(sql, ShouldEqual, `"user"."id" = $1`)
				So(args, ShouldResemble, A{&id})
			})
		})
		Convey("ByIDPtr (required and zero)", func() {
			Convey("nil (should error)", func() {
				_, _, err := compose(byIDPtr(nil).RequiredZero())
				So(err, ShouldBeError, "missing id")
			})
			Convey("zero", func() {
				var id int64 = 0
				sql, args, err := compose(byIDPtr(&id).RequiredZero())
				So(err, ShouldBeNil)
				So(sql, ShouldEqual, `"user"."id" = $1`)
				So(args, ShouldResemble, A{&id})
			})
			Convey("not empty", func() {
				var id int64 = 123
				sql, args, err := compose(byIDPtr(&id).RequiredZero())
				So(err, ShouldBeNil)
				So(sql, ShouldEqual, `"user"."id" = $1`)
				So(args, ShouldResemble, A{&id})
			})
		})
		Convey("ByIDPtr (required and null)", func() {
			Convey("nil (should error)", func() {
				_, _, err := compose(byIDPtr(nil).RequiredNull())
				So(err, ShouldBeError, "missing id")
			})
			Convey("zero (should be null)", func() {
				var id int64 = 0
				sql, args, err := compose(byIDPtr(&id).RequiredNull())
				So(err, ShouldBeNil)
				So(sql, ShouldEqual, `"user"."id" IS NULL`)
				So(args, ShouldResemble, A{})
			})
			Convey("not empty", func() {
				var id int64 = 123
				sql, args, err := compose(byIDPtr(&id).RequiredNull())
				So(err, ShouldBeNil)
				So(sql, ShouldEqual, `"user"."id" = $1`)
				So(args, ShouldResemble, A{&id})
			})
		})
		Convey("ByIDPtr (optional and zero)", func() {
			Convey("nil (should skip)", func() {
				sql, args, err := compose(byIDPtr(nil).OptionalZero())
				So(err, ShouldBeNil)
				So(sql, ShouldEqual, ``)
				So(args, ShouldResemble, A{})
			})
			Convey("zero", func() {
				var id int64 = 0
				sql, args, err := compose(byIDPtr(&id).OptionalZero())
				So(err, ShouldBeNil)
				So(sql, ShouldEqual, `"user"."id" = $1`)
				So(args, ShouldResemble, A{&id})
			})
			Convey("not empty", func() {
				var id int64 = 123
				sql, args, err := compose(byIDPtr(&id).OptionalZero())
				So(err, ShouldBeNil)
				So(sql, ShouldEqual, `"user"."id" = $1`)
				So(args, ShouldResemble, A{&id})
			})
		})
		Convey("ByIDPtr (optional and null)", func() {
			Convey("nil (should skip)", func() {
				sql, args, err := compose(byIDPtr(nil).OptionalNull())
				So(err, ShouldBeNil)
				So(sql, ShouldEqual, ``)
				So(args, ShouldResemble, A{})
			})
			Convey("zero", func() {
				var id int64 = 0
				sql, args, err := compose(byIDPtr(&id).OptionalNull())
				So(err, ShouldBeNil)
				So(sql, ShouldEqual, `"user"."id" IS NULL`)
				So(args, ShouldResemble, A{})
			})
			Convey("not empty", func() {
				var id int64 = 123
				sql, args, err := compose(byIDPtr(&id).OptionalNull())
				So(err, ShouldBeNil)
				So(sql, ShouldEqual, `"user"."id" = $1`)
				So(args, ShouldResemble, A{&id})
			})
		})
	})
}
