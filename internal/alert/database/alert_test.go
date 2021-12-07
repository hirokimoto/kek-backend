package database

import (
	accountDB "kek-backend/internal/account/database"
	accountModel "kek-backend/internal/account/model"
	"kek-backend/internal/alert/model"
	"kek-backend/internal/database"
	"kek-backend/pkg/logging"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
)

var dUser = accountModel.Account{
	Username: "user1",
	Email:    "user1@gmail.com",
	Password: "password",
}

type DBSuite struct {
	suite.Suite
	db        AlertDB
	accountDB accountDB.AccountDB
	originDB  *gorm.DB
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(DBSuite))
}

func (s *DBSuite) SetupSuite() {
	logging.SetLevel(zapcore.FatalLevel)
	s.originDB = database.NewTestDatabase(s.T(), true)
	s.db = &alertDB{db: s.originDB}
	s.accountDB = accountDB.NewAccountDB(s.originDB)
}

func (s *DBSuite) SetupTest() {
	s.NoError(database.DeleteRecordAll(s.T(), s.originDB, []string{
		"comments", "id > 0",
		"alert_tags", "alert_id > 0",
		"tags", "id > 0",
		"alerts", "id > 0",
		"accounts", "id > 0",
	}))
	s.NoError(s.accountDB.Save(nil, &dUser))
}

func (s *DBSuite) TestSaveAlert() {
	// given
	s.NoError(s.originDB.Create(&model.Tag{Name: "tag1"}).Error)
	alert := newAlert("title1", "title1", "body", dUser, []string{"tag1", "tag2"})

	// when
	now := time.Now()
	err := s.db.SaveAlert(nil, alert)

	// then
	s.NoError(err)
	find, err := s.db.FindAlertBySlug(nil, alert.Slug)
	s.NoError(err)
	s.NotEqual(0, find.ID)
	s.Equal(alert.Slug, find.Slug)
	s.Equal(alert.Title, find.Title)
	s.Equal(alert.Body, find.Body)
	s.WithinDuration(now, find.CreatedAt, time.Second)
	s.WithinDuration(now, find.UpdatedAt, time.Second)
	s.Equal(int64(0), find.DeletedAtUnix)
	s.Equal(alert.Author, dUser)
	s.assertAlertTag(find, []string{"tag1", "tag2"})
}

func (s *DBSuite) TestSaveAlert_WithSameSlugAfterDeleted() {
	// given
	alert := newAlert("title1", "title1", "body", dUser, []string{"tag1", "tag2"})
	s.NoError(s.db.SaveAlert(nil, alert))
	s.NoError(s.db.DeleteAlertBySlug(nil, dUser.ID, alert.Slug))

	alert2 := newAlert(alert.Slug, alert.Title, alert.Body, dUser, []string{})

	// when
	err := s.db.SaveAlert(nil, alert2)

	// then
	s.NoError(err)
}

func (s *DBSuite) TestSaveAlert_FailIfDuplicateSlug() {
	// given
	alert := newAlert("title1", "title1", "body", dUser, []string{"tag1"})
	s.NoError(s.db.SaveAlert(nil, alert))

	// when
	alert2 := newAlert(alert.Slug, alert.Title, alert.Body, dUser, nil)
	err := s.db.SaveAlert(nil, alert2)

	// then
	s.Error(err)
	s.Equal(database.ErrKeyConflict, err)
}

func (s *DBSuite) TestFindAlertBySlug() {
	// given
	alert := newAlert("title1", "title1", "body", dUser, []string{"tag1"})
	//now := time.Now()
	s.NoError(s.db.SaveAlert(nil, alert))

	// when
	find, err := s.db.FindAlertBySlug(nil, alert.Slug)

	// then
	s.NoError(err)
	s.assertAlert(alert, find)
}

func (s *DBSuite) TestFindAlertBySlug_FailIfNotExist() {
	// when
	find, err := s.db.FindAlertBySlug(nil, "not-exist-slug")

	// then
	s.Nil(find)
	s.Error(err)
	s.Equal(database.ErrNotFound, err)
}

func (s *DBSuite) TestFindAlertBySlug_FailIfDeleted() {
	// given
	alert := newAlert("title1", "title1", "body", dUser, []string{"tag1"})
	s.NoError(s.db.SaveAlert(nil, alert))
	_, err := s.db.FindAlertBySlug(nil, alert.Slug)
	s.NoError(err)
	s.NoError(s.db.DeleteAlertBySlug(nil, dUser.ID, alert.Slug))

	// when
	find, err := s.db.FindAlertBySlug(nil, alert.Slug)

	// then
	s.Nil(find)
	s.Error(err)
	s.Equal(database.ErrNotFound, err)
}

func (s *DBSuite) TestFindAlerts() {
	// given
	// User1
	// alert1 - tag1, tag2 <- second itr [0]
	// alert2 - tag1		 <- first itr [1]
	// alert3 - tag4
	// alert4 - tag3
	// alert5 - tag1		 <- first itr [0]
	// alert6 - tag1 (deleted)
	// User2
	// alert7 - tag1
	user1 := accountModel.Account{Username: "test-user1", Email: "test-user1@gmail.com", Password: "password"}
	s.NoError(s.accountDB.Save(nil, &user1))
	alert1 := newAlert("alert1", "alert1", "body1", user1, []string{"tag1", "tag2"})
	s.NoError(s.db.SaveAlert(nil, alert1))
	alert2 := newAlert("alert2", "alert2", "body2", user1, []string{"tag1"})
	s.NoError(s.db.SaveAlert(nil, alert2))
	alert3 := newAlert("alert3", "alert3", "body3", user1, []string{"tag4"})
	s.NoError(s.db.SaveAlert(nil, alert3))
	alert4 := newAlert("alert4", "alert4", "body4", user1, []string{"tag3"})
	s.NoError(s.db.SaveAlert(nil, alert4))
	alert5 := newAlert("alert5", "alert5", "body5", user1, []string{"tag1"})
	s.NoError(s.db.SaveAlert(nil, alert5))
	alert6 := newAlert("alert6", "alert6", "body6", user1, []string{"tag1"})
	s.NoError(s.db.SaveAlert(nil, alert6))
	s.NoError(s.db.DeleteAlertBySlug(nil, user1.ID, alert6.Slug))

	user2 := accountModel.Account{Username: "test-user2", Email: "test-user2@gmail.com", Password: "password"}
	s.NoError(s.accountDB.Save(nil, &user2))
	alert7 := newAlert("alert7", "alert7", "body7", user2, []string{"tag1"})
	s.NoError(s.db.SaveAlert(nil, alert7))

	criteria := IterateAlertCriteria{
		Tags:   []string{"tag1", "tag2"},
		Author: user1.Username,
		Offset: 0,
		Limit:  2,
	}

	// when : first iteration
	results, total, err := s.db.FindAlerts(nil, criteria)

	// then
	s.NoError(err)
	s.Equal(int64(3), total)
	s.Equal(2, len(results))
	s.assertAlert(alert5, results[0])
	s.assertAlert(alert2, results[1])

	// second iteration
	criteria.Offset = criteria.Offset + uint(len(results))
	results, total, err = s.db.FindAlerts(nil, criteria)

	// then
	s.NoError(err)
	s.Equal(int64(3), total)
	s.Equal(1, len(results))
	s.assertAlert(alert1, results[0])
}

func (s *DBSuite) TestDeleteAlertBySlug() {
	// given
	alert := newAlert("title1", "title1", "body", dUser, []string{"tag1"})
	s.NoError(s.db.SaveAlert(nil, alert))

	// when
	err := s.db.DeleteAlertBySlug(nil, dUser.ID, alert.Slug)

	// then
	s.NoError(err)
	find, err := s.db.FindAlertBySlug(nil, alert.Slug)
	s.Nil(find)
	s.Equal(database.ErrNotFound, err)
}

func (s *DBSuite) TestDeleteAlertBySlug_FailIfNotExist() {
	// given
	alert := newAlert("title1", "title1", "body", dUser, []string{"tag1"})
	s.NoError(s.db.SaveAlert(nil, alert))
	alert2 := newAlert("title2", "title2", "body", dUser, []string{"tag1"})
	s.NoError(s.db.SaveAlert(nil, alert2))
	s.NoError(s.db.DeleteAlertBySlug(nil, alert2.Author.ID, alert2.Slug))

	cases := []struct {
		AuthorID uint
		Slug     string
	}{
		{
			AuthorID: dUser.ID,
			Slug:     "not-exist-slug",
		}, {
			AuthorID: dUser.ID + 1000,
			Slug:     alert.Slug,
		}, {
			AuthorID: dUser.ID,
			Slug:     alert2.Slug,
		},
	}

	for _, tc := range cases {
		// when
		err := s.db.DeleteAlertBySlug(nil, tc.AuthorID, tc.Slug)

		// then
		s.Error(err)
		s.Equal(database.ErrNotFound, err)
	}
}

func (s *DBSuite) assertAlert(expected, actual *model.Alert) {
	s.Equal(expected.Slug, actual.Slug)
	s.Equal(expected.Title, actual.Title)
	s.Equal(expected.Body, actual.Body)
	s.WithinDuration(expected.CreatedAt, actual.CreatedAt, time.Second)
	s.WithinDuration(expected.UpdatedAt, actual.UpdatedAt, time.Second)
	s.Equal(expected.DeletedAtUnix, actual.DeletedAtUnix)
	s.Equal(expected.Author.ID, actual.Author.ID)
	s.Equal(expected.Author.Email, actual.Author.Email)
	s.Equal(expected.Author.Username, actual.Author.Username)
	var tags []string
	for _, tag := range expected.Tags {
		tags = append(tags, tag.Name)
	}
	s.assertAlertTag(actual, tags)
}

func (s *DBSuite) assertAlertTag(alert *model.Alert, tags []string) {
	s.Equal(len(alert.Tags), len(tags))
	if len(alert.Tags) == 0 {
		return
	}
	m := make(map[string]struct{})
	for _, tag := range alert.Tags {
		m[tag.Name] = struct{}{}
	}
	for _, tag := range tags {
		_, ok := m[tag]
		s.True(ok)
	}
}

func newAlert(slug, title, body string, author accountModel.Account, tags []string) *model.Alert {
	var tagArr []*model.Tag
	for _, tag := range tags {
		tagArr = append(tagArr, &model.Tag{Name: tag})
	}
	return &model.Alert{
		Slug:   slug,
		Title:  title,
		Body:   body,
		Author: author,
		Tags:   tagArr,
	}
}
