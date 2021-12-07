package database

import (
	"context"
	"fmt"
	"kek-backend/internal/alert/model"
	"kek-backend/internal/database"
	"kek-backend/pkg/logging"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type IterateAlertCriteria struct {
	Tags   []string
	Author string
	Offset uint
	Limit  uint
}

//go:generate mockery --name AlertDB --filename alert_mock.go
type AlertDB interface {
	RunInTx(ctx context.Context, f func(ctx context.Context) error) error

	// SaveAlert saves a given alert with tags.
	// if not exist tags, then save a new tag
	SaveAlert(ctx context.Context, alert *model.Alert) error

	// FindAlertBySlug returns a alert with given slug
	// database.ErrNotFound error is returned if not exist
	FindAlertBySlug(ctx context.Context, slug string) (*model.Alert, error)

	// FindAlerts returns alert list with given criteria and total count
	FindAlerts(ctx context.Context, criteria IterateAlertCriteria) ([]*model.Alert, int64, error)

	// DeleteAlertBySlug deletes a alert with given slug
	// and returns nil if success to delete, otherwise returns an error
	DeleteAlertBySlug(ctx context.Context, authorId uint, slug string) error

	// SaveComment saves a comment with given alert slug and comment
	SaveComment(ctx context.Context, slug string, comment *model.Comment) error

	// FindComments returns all comments with given alert slug
	FindComments(ctx context.Context, slug string) ([]*model.Comment, error)

	// DeleteCommentById deletes a comment with given alert slug and comment id
	// database.ErrNotFound error is returned if not exist
	DeleteCommentById(ctx context.Context, authorId uint, slug string, id uint) error

	// DeleteComments deletes all comment with given author id and slug
	// and returns deleted records count
	DeleteComments(ctx context.Context, authorId uint, slug string) (int64, error)
}

type alertDB struct {
	db *gorm.DB
}

func (a *alertDB) RunInTx(ctx context.Context, f func(ctx context.Context) error) error {
	tx := a.db.Begin()
	if tx.Error != nil {
		return errors.Wrap(tx.Error, "start tx")
	}

	ctx = database.WithDB(ctx, tx)
	if err := f(ctx); err != nil {
		if err1 := tx.Rollback().Error; err1 != nil {
			return errors.Wrap(err, fmt.Sprintf("rollback tx: %v", err1.Error()))
		}
		return errors.Wrap(err, "invoke function")
	}
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit tx: %v", err)
	}
	return nil
}

func (a *alertDB) SaveAlert(ctx context.Context, alert *model.Alert) error {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, a.db)
	logger.Debugw("alert.db.SaveAlert", "alert", alert)

	// TODO : transaction
	for _, tag := range alert.Tags {
		if err := db.WithContext(ctx).FirstOrCreate(&tag, "name = ?", tag.Name).Error; err != nil {
			logger.Errorw("alert.db.SaveAlert failed to first or save tag", "err", err)
			return err
		}
	}

	if err := db.WithContext(ctx).Create(alert).Error; err != nil {
		logger.Errorw("alert.db.SaveAlert failed to save alert", "err", err)
		if database.IsKeyConflictErr(err) {
			return database.ErrKeyConflict
		}
		return err
	}
	return nil
}

func (a *alertDB) FindAlertBySlug(ctx context.Context, slug string) (*model.Alert, error) {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, a.db)
	logger.Debugw("alert.db.FindAlertBySlug", "slug", slug)

	var ret model.Alert
	// 1) load alert with author
	// SELECT alerts.*, accounts.*
	// FROM `alerts` LEFT JOIN `accounts` `Author` ON `alerts`.`author_id` = `Author`.`id`
	// WHERE slug = "title1" AND deleted_at_unix = 0 ORDER BY `alerts`.`id` LIMIT 1
	err := db.WithContext(ctx).Joins("Author").
		First(&ret, "slug = ? AND deleted_at_unix = 0", slug).Error
	// 2) load tags
	if err == nil {
		// SELECT * from tags JOIN alert_tags ON alert_tags.tag_id = tags.id AND alert_tags.alert_id = ?
		err = db.WithContext(ctx).Model(&ret).Association("Tags").Find(&ret.Tags)
	}

	if err != nil {
		logger.Errorw("failed to find alert", "err", err)
		if database.IsRecordNotFoundErr(err) {
			return nil, database.ErrNotFound
		}
		return nil, err
	}
	return &ret, nil
}

func (a *alertDB) FindAlerts(ctx context.Context, criteria IterateAlertCriteria) ([]*model.Alert, int64, error) {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, a.db)
	logger.Debugw("alert.db.FindAlerts", "criteria", criteria)

	chain := db.WithContext(ctx).Table("alerts a").Where("deleted_at_unix = 0")
	if len(criteria.Tags) != 0 {
		chain = chain.Where("t.name IN ?", criteria.Tags)
	}
	if criteria.Author != "" {
		chain = chain.Where("au.username = ?", criteria.Author)
	}
	if len(criteria.Tags) != 0 {
		chain = chain.Joins("LEFT JOIN alert_tags ats on ats.alert_id = a.id").
			Joins("LEFT JOIN tags t on t.id = ats.tag_id")
	}
	if criteria.Author != "" {
		chain = chain.Joins("LEFT JOIN accounts au on au.id = a.author_id")
	}

	// get total count
	var totalCount int64
	err := chain.Distinct("a.id").Count(&totalCount).Error
	if err != nil {
		logger.Error("failed to get total count", "err", err)
	}

	// get alert ids
	rows, err := chain.Select("DISTINCT(a.id) id").
		Offset(int(criteria.Offset)).
		Limit(int(criteria.Limit)).
		Order("a.id DESC").
		Rows()
	if err != nil {
		logger.Error("failed to read alert ids", "err", err)
		return nil, 0, err
	}
	var ids []uint
	for rows.Next() {
		var id uint
		err := rows.Scan(&id)
		if err != nil {
			logger.Error("failed to scan id from id rows", "err", err)
			return nil, 0, err
		}
		ids = append(ids, id)
	}

	// get alerts with author by ids
	var ret []*model.Alert
	if len(ids) == 0 {
		return []*model.Alert{}, totalCount, nil
	}
	err = db.WithContext(ctx).Joins("Author").
		Where("alerts.id IN (?)", ids).
		Order("alerts.id DESC").
		Find(&ret).Error
	if err != nil {
		logger.Error("failed to find alert by ids", "err", err)
		return nil, 0, err
	}

	// get tags by alert ids
	ma := make(map[uint]*model.Alert)
	for _, r := range ret {
		ma[r.ID] = r
	}
	type AlertTag struct {
		model.Tag
		AlertId uint
	}
	batchSize := 100 // TODO : config
	for i := 0; i < len(ret); i += batchSize {
		var at []*AlertTag
		last := i + batchSize
		if last > len(ret) {
			last = len(ret)
		}

		err = db.WithContext(ctx).Table("tags").
			Where("alert_tags.alert_id IN (?)", ids[i:last]).
			Joins("LEFT JOIN alert_tags ON alert_tags.tag_id = tags.id").
			Select("tags.*, alert_tags.alert_id alert_id").
			Find(&at).Error

		if err != nil {
			logger.Error("failed to load tags by alert ids", "alertIds", ids[i:last], "err", err)
			return nil, 0, err
		}
		for _, tag := range at {
			a := ma[tag.AlertId]
			a.Tags = append(a.Tags, &tag.Tag)
		}
	}
	return ret, totalCount, nil
}

func (a *alertDB) DeleteAlertBySlug(ctx context.Context, authorId uint, slug string) error {
	logger := logging.FromContext(ctx)
	db := database.FromContext(ctx, a.db)
	logger.Debugw("alert.db.DeleteAlertBySlug", "slug", slug)

	// delete alert
	chain := db.WithContext(ctx).Model(&model.Alert{}).
		Where("slug = ? AND deleted_at_unix = 0", slug).
		Where("author_id = ?", authorId).
		Update("deleted_at_unix", time.Now().Unix())
	if chain.Error != nil {
		logger.Errorw("failed to delete an alert", "err", chain.Error)
		return chain.Error
	}
	if chain.RowsAffected == 0 {
		logger.Error("failed to delete an alert because not found")
		return database.ErrNotFound
	}
	// delete alert tag relation
	query := `DELETE ats FROM alert_tags ats
		LEFT JOIN alerts a on a.id = ats.alert_id
		WHERE a.slug = ?;`
	if err := db.WithContext(ctx).Exec(query, slug).Error; err != nil {
		logger.Errorw("failed to delete relation of alerts and tags", "err", err)
		return err
	}
	return nil
}

// NewAlertDB creates a new alert db with given db
func NewAlertDB(db *gorm.DB) AlertDB {
	return &alertDB{
		db: db,
	}
}
