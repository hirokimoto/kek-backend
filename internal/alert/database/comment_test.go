package database

import (
	"kek-backend/internal/alert/model"
	"kek-backend/internal/database"
	"time"
)

func (s *DBSuite) TestSaveComment() {
	// given
	alert := newAlert("title1", "title1", "body", dUser, []string{"tag1", "tag2"})
	s.NoError(s.db.SaveAlert(nil, alert))

	comment := model.Comment{Body: "comment1", Author: dUser}

	// when
	err := s.db.SaveComment(nil, alert.Slug, &comment)

	// then
	s.NoError(err)
}

func (s *DBSuite) TestSaveComment_FailIfNotExistOrDeleted() {
	alert := newAlert("title1", "title1", "body", dUser, []string{"tag1", "tag2"})
	s.NoError(s.db.SaveAlert(nil, alert))
	s.NoError(s.db.DeleteAlertBySlug(nil, dUser.ID, alert.Slug))

	cases := []struct {
		Slug string
	}{
		{
			Slug: "not-exist",
		}, {
			Slug: alert.Slug,
		},
	}

	for _, tc := range cases {
		err := s.db.SaveComment(nil, tc.Slug, &model.Comment{Body: "comment"})
		s.Error(err)
		s.Equal(database.ErrNotFound, err)
	}
}

func (s *DBSuite) TestFindComments() {
	// given
	alert := newAlert("title1", "title1", "body", dUser, []string{"tag1", "tag2"})
	s.NoError(s.db.SaveAlert(nil, alert))
	c1 := model.Comment{Body: "comment1", Author: dUser}
	s.NoError(s.db.SaveComment(nil, alert.Slug, &c1))
	c2 := model.Comment{Body: "comment2", Author: dUser}
	s.NoError(s.db.SaveComment(nil, alert.Slug, &c2))
	c3 := model.Comment{Body: "comment3", Author: dUser}
	s.NoError(s.db.SaveComment(nil, alert.Slug, &c3))
	s.NoError(s.db.DeleteCommentById(nil, dUser.ID, alert.Slug, c3.ID))

	// when
	comments, err := s.db.FindComments(nil, alert.Slug)

	// then
	s.NoError(err)
	s.Equal(2, len(comments))
	s.assertAlertComment(&c2, comments[0])
	s.assertAlertComment(&c1, comments[1])
}

func (s *DBSuite) TestDeleteCommentById() {
	// given
	alert := newAlert("title1", "title1", "body", dUser, []string{"tag1", "tag2"})
	s.NoError(s.db.SaveAlert(nil, alert))
	c := model.Comment{Body: "comment1", Author: dUser}
	s.NoError(s.db.SaveComment(nil, alert.Slug, &c))

	// when
	err := s.db.DeleteCommentById(nil, dUser.ID, alert.Slug, c.ID)

	// then
	s.NoError(err)
	find, err := s.db.FindComments(nil, alert.Slug)
	s.NoError(err)
	s.Empty(find)
}

func (s *DBSuite) TestDeleteCommentById_FailIfNotExist() {
	alert := newAlert("title1", "title1", "body", dUser, []string{"tag1", "tag2"})
	s.NoError(s.db.SaveAlert(nil, alert))
	c1 := model.Comment{Body: "comment1", Author: dUser}
	s.NoError(s.db.SaveComment(nil, alert.Slug, &c1))
	c2 := model.Comment{Body: "comment2", Author: dUser}
	s.NoError(s.db.SaveComment(nil, alert.Slug, &c2))
	s.NoError(s.db.DeleteCommentById(nil, dUser.ID, alert.Slug, c2.ID))

	cases := []struct {
		Name     string
		AuthorID uint
		Slug     string
		Id       uint
	}{
		{
			Name:     "another author id",
			AuthorID: dUser.ID + 1,
			Slug:     alert.Slug,
			Id:       c1.ID,
		}, {
			Name:     "another slug",
			AuthorID: dUser.ID,
			Slug:     alert.Slug + "-not-exist",
			Id:       c1.ID,
		}, {
			Name:     "already deleted",
			AuthorID: dUser.ID,
			Slug:     alert.Slug,
			Id:       c2.ID,
		},
	}

	for _, tc := range cases {
		// when
		err := s.db.DeleteCommentById(nil, tc.AuthorID, tc.Slug, tc.Id)
		// then
		s.Error(err)
		s.Equal(database.ErrNotFound, err)
	}
}

func (s *DBSuite) TestDeleteComments() {
	// given
	alert := newAlert("title1", "title1", "body", dUser, []string{"tag1", "tag2"})
	s.NoError(s.db.SaveAlert(nil, alert))
	c1 := model.Comment{Body: "comment1", Author: dUser}
	s.NoError(s.db.SaveComment(nil, alert.Slug, &c1))
	c2 := model.Comment{Body: "comment2", Author: dUser}
	s.NoError(s.db.SaveComment(nil, alert.Slug, &c2))
	c3 := model.Comment{Body: "comment3", Author: dUser}
	s.NoError(s.db.SaveComment(nil, alert.Slug, &c3))
	s.NoError(s.db.DeleteCommentById(nil, dUser.ID, alert.Slug, c3.ID))

	// when
	deleted, err := s.db.DeleteComments(nil, dUser.ID, alert.Slug)

	// then
	s.NoError(err)
	s.Equal(int64(2), deleted)
	find, err := s.db.FindComments(nil, alert.Slug)
	s.NoError(err)
	s.Empty(find)
}

func (s *DBSuite) assertAlertComment(expected, actual *model.Comment) {
	if expected == nil && actual == nil {
		return
	}
	s.NotNil(expected)
	s.NotNil(actual)

	s.Equal(expected.ID, actual.ID)
	s.Equal(expected.Body, actual.Body)
	s.Equal(expected.Slug, actual.Slug)
	s.Equal(expected.Author.Username, actual.Author.Username)
	s.Equal(expected.Author.Email, actual.Author.Email)
	s.Equal(expected.Author.Bio, actual.Author.Bio)
	s.Equal(expected.Author.Image, actual.Author.Image)
	s.WithinDuration(expected.CreatedAt, actual.CreatedAt, time.Second)
	s.WithinDuration(expected.UpdatedAt, actual.UpdatedAt, time.Second)
	s.Equal(expected.DeletedAt, actual.DeletedAt)
}
