package database

import (
	"context"
	"kek-backend/internal/alert/model"
	"kek-backend/internal/database"
	"kek-backend/pkg/logging"

	"gorm.io/gorm"
)

func (a *alertDB) SaveComment(ctx context.Context, slug string, comment *model.Comment) error {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, a.db)
	logger.Debugw("alert.db.SaveComment", "slug", slug, "comment", comment)

	var alertCount int64
	err := db.Model(&model.Alert{}).Where("slug = ? AND deleted_at_unix = 0", slug).Count(&alertCount).Error
	if err == nil && alertCount == 0 {
		err = gorm.ErrRecordNotFound
	}
	if err != nil {
		logger.Errorw("alert.db.SaveComment failed to find a alert", "err", err)
		if database.IsRecordNotFoundErr(err) {
			return database.ErrNotFound
		}
		return err
	}

	comment.Slug = slug
	if err := db.WithContext(ctx).Create(comment).Error; err != nil {
		logger.Errorw("alert.db.SaveComment failed to save comment", "err", err)
		return err
	}
	return nil
}

func (a *alertDB) FindComments(ctx context.Context, slug string) ([]*model.Comment, error) {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, a.db)
	logger.Debugw("alert.db.FindComments", "slug", slug)

	var ret []*model.Comment
	err := db.Joins("Author").
		Where("slug = ? AND deleted_at IS NULL", slug).
		Order("id DESC").
		Find(&ret).Error
	if err != nil {
		logger.Errorw("alert.db.FindComments failed to find comments", "err", err)
		return nil, err
	}
	return ret, nil
}

func (a *alertDB) DeleteCommentById(ctx context.Context, authorId uint, slug string, id uint) error {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, a.db)
	logger.Debugw("alert.db.DeleteCommentById", "authorId", authorId, "slug", slug, "id", id)

	chain := db.Where("author_id = ?", authorId).
		Where("slug = ?", slug).
		Where("id = ? AND deleted_at IS NULL", id).
		Delete(&model.Comment{})

	if chain.Error != nil {
		logger.Errorw("alert.db.DeleteCommentById failed to delete a comment", "err", chain.Error)
		return chain.Error
	}
	if chain.RowsAffected == 0 {
		logger.Error("alert.db.DeleteCommentById empty rows affected")
		return database.ErrNotFound
	}
	return nil
}

func (a *alertDB) DeleteComments(ctx context.Context, authorId uint, slug string) (int64, error) {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, a.db)
	logger.Debugw("alert.db.DeleteComments", "authorId", authorId, "slug", slug)

	chain := db.Where("author_id = ?", authorId).
		Where("slug = ? AND deleted_at IS NULL", slug).
		Delete(&model.Comment{})
	if chain.Error != nil {
		logger.Errorw("alert.db.DeleteComments failed to delete comments", "err", chain.Error)
		return 0, chain.Error
	}
	return chain.RowsAffected, nil
}
