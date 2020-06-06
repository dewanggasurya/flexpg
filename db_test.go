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
					err := cur.Fetchs(&ms, 0)
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
				err := cur.Fetchs(&ms, 0)
				cv.So(err, cv.ShouldBeNil)

				toolkit.Logger().Infof("\nResults:\n%s\n", toolkit.JsonString(ms))
			})
		})
	})
}

type TestData struct {
	ID      string
	Title   string
	DataDec float64
	Created time.Time
}
