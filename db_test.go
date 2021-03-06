package flexpg_test

import (
	"errors"
	"testing"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"github.com/eaciit/toolkit"
	"github.com/smartystreets/goconvey/convey"
	cv "github.com/smartystreets/goconvey/convey"
)

var (
	connString = "postgres://localhost/testdb?sslmode=disable&binary_parameters=yes"
	tableName  = "testable"
)

func connect() (dbflex.IConnection, error) {
	dbflex.Logger().SetLevelStdOut(toolkit.ErrorLevel, true)
	dbflex.Logger().SetLevelStdOut(toolkit.InfoLevel, true)
	dbflex.Logger().SetLevelStdOut(toolkit.WarningLevel, true)
	dbflex.Logger().SetLevelStdOut(toolkit.DebugLevel, true)

	conn, err := dbflex.NewConnectionFromURI(connString, nil)
	if err != nil {
		return nil, errors.New("unable to connect. " + err.Error())
	}
	err = conn.Connect()
	if err != nil {
		return nil, errors.New("unable to connect. " + err.Error())
	}
	return conn, nil
}

func TestMigration(t *testing.T) {
	cv.Convey("connecting", t, func() {
		conn, err := connect()
		cv.So(err, cv.ShouldBeNil)
		defer conn.Close()

		cv.Convey("migrate", func() {
			if conn.HasTable(tableName) {
				err = conn.DropTable(tableName)
				cv.So(err, cv.ShouldBeNil)
			}

			err = conn.EnsureTable(tableName, []string{"ID"}, new(TestData))
			cv.So(err, cv.ShouldBeNil)

			cv.Convey("validate ", func() {
				has := conn.HasTable(tableName)
				cv.So(has, cv.ShouldBeTrue)

				cv.Convey("migrate update", func() {
					err = conn.EnsureTable(tableName, []string{"ID"}, new(TestDataNew))
					cv.So(err, cv.ShouldBeNil)

					cv.Convey("validate update", func() {
						sql := "select column_name from information_schema.columns where table_name='" + tableName + "'"
						fields := []toolkit.M{}
						err = conn.Cursor(dbflex.SQL(sql), nil).Fetchs(&fields, 0).Close()
						cv.So(err, cv.ShouldBeNil)
						cv.Printf("\nFields: %s\n", toolkit.JsonString(fields))
						cv.So(len(fields), cv.ShouldEqual, 15)
					})
				})
			})
		})
	})
}

func TestQueryM(t *testing.T) {
	cv.Convey("connecting", t, func() {
		conn, err := connect()
		cv.So(err, cv.ShouldBeNil)
		defer conn.Close()

		convey.Convey("insert", func() {
			data := toolkit.M{}.
				Set("id", "TestData1").
				Set("title", "Title aja lah").
				Set("datadec", 80.32).
				Set("category", Inbound).
				Set("status", Received).
				Set("tags", []string{"a", "b", "c", "d"}).
				Set("sequences", []int64{9, 11, 13, 15, 17}).
				Set("datainterfaces", []interface{}{"hola", 1, 2, "como", "estas", 3}).
				Set("dataint64", 9000).
				Set("meta", TestObject{
					ID:          "Test",
					CreatedDate: time.Now(),
					Name:        "metadata",
					Tags:        []string{"z", "y", "x", "w"},
				}).
				Set("created", time.Now())

			cmd := dbflex.From(tableName).Insert()
			_, e := conn.Execute(cmd, toolkit.M{}.Set("data", data))
			cv.So(e, convey.ShouldBeNil)

			cv.Convey("querying", func() {
				cmd = dbflex.From(tableName).Select()
				cur := conn.Cursor(cmd, nil)
				cv.So(cur.Error(), cv.ShouldBeNil)

				cv.Convey("get results", func() {
					ms := []toolkit.M{}
					err := cur.Fetchs(&ms, 0).Close()
					cv.So(err, cv.ShouldBeNil)
					cv.So(len(ms), cv.ShouldBeGreaterThan, 0)

					toolkit.Logger().Infof("\nResults:\n%s\n", toolkit.JsonString(ms))
				})
			})
		})

	})
}

func TestQueryObj(t *testing.T) {
	cv.Convey("connecting", t, func() {
		conn, err := connect()
		cv.So(err, cv.ShouldBeNil)
		defer conn.Close()

		cv.Convey("querying", func() {
			cmd := dbflex.From(tableName).Select()
			cur := conn.Cursor(cmd, nil)
			cv.So(cur.Error(), cv.ShouldBeNil)

			cv.Convey("get results", func() {
				ms := []struct {
					ID      string
					Title   string
					DataDec float64
					Created time.Time
				}{}
				err := cur.Fetchs(&ms, 0).Close()
				cv.So(err, cv.ShouldBeNil)

				toolkit.Logger().Infof("\nResults:\n%s\n", toolkit.JsonString(ms))
			})
		})
	})
}

func TestDate(t *testing.T) {
	cv.Convey("connecting", t, func() {
		conn, err := connect()
		cv.So(err, cv.ShouldBeNil)
		defer conn.Close()

		cv.Convey("save date data", func() {
			cmd := dbflex.From(tableName).Insert()
			obj := new(TestData)
			obj.ID = "date1"
			obj.Title = "Date 1"
			obj.DataDec = 305
			obj.Created = toolkit.String2Date("01-Apr-1980 00:00:00", "dd-MMM-yyyy HH:mm:ss")
			if _, err = conn.Execute(cmd, toolkit.M{}.Set("data", obj)); err != nil {
				cmd = dbflex.From(tableName).Update().Where(dbflex.Eq("ID", "date1"))
				_, err = conn.Execute(cmd, toolkit.M{}.Set("data", obj))
			}
			cv.So(err, cv.ShouldBeNil)

			cv.Convey("compare equal date", func() {
				ms := []struct {
					ID      string
					Title   string
					DataDec float64
					Created time.Time
				}{}
				cmd = dbflex.From(tableName).Select().Where(dbflex.Eq("created", toolkit.String2Date("01-Apr-1980", "dd-MMM-yyyy")))
				err := conn.Cursor(cmd, nil).Fetchs(&ms, 0).Close()
				cv.So(err, cv.ShouldBeNil)
				cv.So(len(ms), cv.ShouldBeGreaterThan, 0)

				toolkit.Logger().Infof("\nResults:\n%s\n", toolkit.JsonString(ms))
			})
		})
	})
}

func TestRollback(t *testing.T) {
	cv.Convey("connecting", t, func() {
		conn, err := connect()
		cv.So(err, cv.ShouldBeNil)
		defer conn.Close()

		cv.Convey("save date data", func() {
			conn.BeginTx()
			cmd := dbflex.From(tableName).Insert()
			obj := new(TestData)
			obj.ID = "tx1"
			obj.Title = "tx1"
			obj.DataDec = 305
			obj.Created = toolkit.String2Date("01-Apr-1980 00:00:00", "dd-MMM-yyyy HH:mm:ss")
			if _, err = conn.Execute(cmd, toolkit.M{}.Set("data", obj)); err != nil {
				cmd = dbflex.From(tableName).Update().Where(dbflex.Eq("ID", "tx1"))
				_, err = conn.Execute(cmd, toolkit.M{}.Set("data", obj))
			}
			cv.So(err, cv.ShouldBeNil)

			cv.Convey("get data", func() {
				ms := []struct {
					ID      string
					Title   string
					DataDec float64
					Created time.Time
				}{}
				cmd = dbflex.From(tableName).Select().Where(dbflex.Eq("id", "tx1"))
				err := conn.Cursor(cmd, nil).Fetchs(&ms, 0).Close()
				cv.So(err, cv.ShouldBeNil)
				cv.So(len(ms), cv.ShouldEqual, 1)
				toolkit.Logger().Infof("\nResults:\n%s\n", toolkit.JsonString(ms))

				cv.Convey("rollback", func() {
					err = conn.RollBack()
					cv.So(err, cv.ShouldBeNil)

					cv.Convey("validate", func() {
						cmd = dbflex.From(tableName).Select().Where(dbflex.Eq("id", "tx1"))
						err := conn.Cursor(cmd, nil).Fetchs(&ms, 0).Close()
						cv.So(err, cv.ShouldBeNil)
						cv.So(len(ms), cv.ShouldEqual, 0)
					})
				})
			})
		})
	})
}

func TestCommit(t *testing.T) {
	cv.Convey("connecting", t, func() {
		conn, err := connect()
		cv.So(err, cv.ShouldBeNil)
		defer conn.Close()

		cv.Convey("save date data", func() {
			conn.BeginTx()
			cmd := dbflex.From(tableName).Insert()
			obj := new(TestData)
			obj.ID = "tx-commit"
			obj.Title = "tx-commit"
			obj.DataDec = 305
			obj.Created = toolkit.String2Date("01-Apr-1980 00:00:00", "dd-MMM-yyyy HH:mm:ss")
			if _, err = conn.Execute(cmd, toolkit.M{}.Set("data", obj)); err != nil {
				cmd = dbflex.From(tableName).Update().Where(dbflex.Eq("ID", "tx-commit"))
				_, err = conn.Execute(cmd, toolkit.M{}.Set("data", obj))
			}
			cv.So(err, cv.ShouldBeNil)

			cv.Convey("get data", func() {
				ms := []struct {
					ID      string
					Title   string
					DataDec float64
					Created time.Time
				}{}
				cmd = dbflex.From(tableName).Select().Where(dbflex.Eq("id", "tx-commit"))
				err := conn.Cursor(cmd, nil).Fetchs(&ms, 0).Close()
				cv.So(err, cv.ShouldBeNil)
				cv.So(len(ms), cv.ShouldEqual, 1)
				toolkit.Logger().Infof("\nResults:\n%s\n", toolkit.JsonString(ms))

				cv.Convey("rollback", func() {
					err = conn.Commit()
					cv.So(err, cv.ShouldBeNil)

					cv.Convey("validate", func() {
						cmd = dbflex.From(tableName).Select().Where(dbflex.Eq("id", "tx-commit"))
						err := conn.Cursor(cmd, nil).Fetchs(&ms, 0).Close()
						cv.So(err, cv.ShouldBeNil)
						cv.So(len(ms), cv.ShouldEqual, 1)
					})
				})
			})
		})
	})
}

func TestPopulateSQL(t *testing.T) {
	cv.Convey("connecting", t, func() {
		conn, err := connect()
		cv.So(err, cv.ShouldBeNil)
		defer conn.Close()

		cv.Convey("reading using sql query", func() {
			cmd := dbflex.SQL("select * from " + tableName)
			ms := []toolkit.M{}
			err = conn.Cursor(cmd, nil).Fetchs(&ms, 0).Close()
			cv.So(err, cv.ShouldBeNil)
			cv.So(len(ms), cv.ShouldBeGreaterThan, 0)
			cv.Printf("\ndata returned: %s\n", toolkit.JsonString(ms[:2]))
		})
	})
}

type Status int64

const (
	Pending Status = iota
	Received
	Cancelled
	Rejected
)

type Category string

const (
	Inbound  Category = "inbound"
	Outbound Category = "outbound"
)

type TestData struct {
	ID             string
	Title          string
	Status         Status
	Category       Category
	Tags           []string
	Sequences      []int64
	DataDec        float64
	DataInt        int
	DataInt64      int64
	DataInterfaces []interface{}
	Meta           TestObject
	Metas          []TestObject
	Created        time.Time
	Updated        time.Time
}

type TestObject struct {
	ID          string
	Name        string
	Tags        []string
	CreatedDate time.Time
}

type TestDataNew struct {
	ID      string
	Title   string
	Name    string
	DataDec float64
	DataInt int
	Created time.Time
}
