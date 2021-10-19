package storage

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Video struct {
	ID          int       `gorm:"column:id"`
	UserID      int       `gorm:"column:user_id"`
	Location    string    `gorm:"column:location"`
	URI         string    `gorm:"column:uri"`
	RES         string    `gorm:"column:res"`
	Caption     string    `gorm:"column:caption"`
	Description string    `gorm:"column:description"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

type gormDB struct {
	db *gorm.DB
}

func newGormDB(c *ConnString) (*gormDB, error) {
	db, err := gorm.Open(postgres.Open(composeGormDSN(c)), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open Gorm connection: %w", err)
	}
	return &gormDB{
		db: db,
	}, nil
}

// GetVideosByCaption sends query to the DB and processes the given result
func (g *gormDB) GetVideosByCaption(ctx context.Context, phrase string) ([]*FoundVideo, error) {
	var vids []Video
	req := g.db.
		Select("caption", "uri", "location").
		Where("caption LIKE ?", "%"+phrase+"%").
		Find(&vids)
	if err := req.Error; err != nil {
		return nil, fmt.Errorf("failed to query videos by caption substring: %w", err)
	}
	videos := make([]*FoundVideo, len(vids))
	for i, e := range vids {
		p := &FoundVideo{
			Caption:  e.Caption,
			URI:      e.URI,
			Location: e.Location,
		}
		videos[i] = p
	}
	return videos, nil
}

func (g *gormDB) Close() {}

func composeGormDSN(c *ConnString) string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Europe/Moscow",
		c.Host,
		c.User,
		c.Password,
		c.DBName,
		c.Port,
	)
}
