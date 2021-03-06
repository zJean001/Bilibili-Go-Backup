package wechat

import (
	"context"
	"flag"
	"os"
	"testing"

	"go-common/app/admin/ep/saga/conf"
	"go-common/app/admin/ep/saga/dao"
	"go-common/app/admin/ep/saga/model"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	mydao  *dao.Dao
	wechat *Wechat
	ctx    = context.Background()
)

func TestMain(m *testing.M) {
	var err error
	flag.Set("conf", "../../cmd/saga-admin-test.toml")
	if err = conf.Init(); err != nil {
		panic(err)
	}

	mydao = dao.New()
	defer mydao.Close()
	wechat = New(mydao)

	os.Exit(m.Run())
}

func TestAddRequireVisible(t *testing.T) {
	var (
		err     error
		userMap = make(map[string]model.RequireVisibleUser)
	)

	Convey("TEST addRequireVisible", t, func() {
		err = wechat.addRequireVisible(ctx, "000000")
		So(err, ShouldNotBeNil)
		err = wechat.addRequireVisible(ctx, "001134")
		So(err, ShouldBeNil)
		err = mydao.RequireVisibleUsersRedis(ctx, &userMap)
		So(err, ShouldBeNil)
		So(userMap, ShouldContainKey, "001134")
	})
}

func TestAlreadyInCache(t *testing.T) {
	var (
		err         error
		result      bool
		contactInfo model.ContactInfo
	)

	contactInfo = model.ContactInfo{
		ID:          "111",
		UserName:    "zhangsan",
		UserID:      "222",
		NickName:    "xiaolizi",
		VisibleSaga: true,
	}

	Convey("TEST alreadyInCache", t, func() {
		result, err = wechat.alreadyInCache(ctx, "000")
		So(err, ShouldBeNil)
		So(result, ShouldEqual, false)
		So(mydao.SetRequireVisibleUsersRedis(ctx, &contactInfo), ShouldBeNil)
		result, err = wechat.alreadyInCache(ctx, "222")
		So(err, ShouldBeNil)
		So(result, ShouldEqual, true)
	})
}

func TestSyncContacts(t *testing.T) {
	var (
		err         error
		contactInfo = &model.ContactInfo{
			UserID:   "E10021",
			UserName: "eyotang",
			NickName: "ben大神点C",
		}
		modify = &model.ContactInfo{
			UserID:      "000328",
			UserName:    "eyotang",
			NickName:    "ben大神点C",
			VisibleSaga: false,
		}
		target      *model.ContactInfo
		almostEqual bool
	)

	Convey("TEST sync after add incorrect", t, func() {
		err = wechat.dao.CreateContact(contactInfo)
		So(err, ShouldBeNil)
		target, err = wechat.dao.QueryUserByID(contactInfo.UserID)
		So(err, ShouldBeNil)
		almostEqual = contactInfo.AlmostEqual(target)
		So(almostEqual, ShouldBeTrue)

		err = wechat.SyncContacts(ctx)
		So(err, ShouldBeNil)
		target, err = wechat.dao.QueryUserByID(contactInfo.UserID)
		So(err, ShouldNotBeNil)
	})

	Convey("TEST aync after change", t, func() {
		contactInfo, err = wechat.dao.QueryUserByID(modify.UserID)
		So(err, ShouldBeNil)
		modify.ID = contactInfo.ID
		err = wechat.dao.UptContact(contactInfo)
		So(err, ShouldBeNil)

		err = wechat.SyncContacts(ctx)
		So(err, ShouldBeNil)
		target, err = wechat.dao.QueryUserByID(modify.UserID)
		So(err, ShouldBeNil)
		So(target.VisibleSaga, ShouldBeTrue)
		So(target.UserName, ShouldNotEqual, "eyotang")
	})
}

func TestPushMsg(t *testing.T) {
	var err error
	userName := []string{"wuwei"}
	content := "测试发送企业微信"
	Convey("TEST push message", t, func() {
		err = wechat.PushMsg(ctx, userName, content)
		So(err, ShouldBeNil)
	})
}

func TestAnalysisContacts(t *testing.T) {
	var err error
	Convey("TEST analysis contacts", t, func() {
		err = wechat.AnalysisContacts(ctx)
		So(err, ShouldBeNil)
	})
}
