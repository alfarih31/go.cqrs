// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"time"
)

const TableNameEventStore = "event_store"

// EventStore mapped from table <event_store>
type EventStore struct {
	ID           int64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true" json:"id"`
	EventID      string         `gorm:"column:event_id;type:uuid;not null;uniqueIndex:event_store_event_id_key,priority:1" json:"event_id"`
	EventName    string         `gorm:"column:event_name;type:character varying(255);not null" json:"event_name"`
	EventData    string         `gorm:"column:event_data;type:text;not null" json:"event_data"`
	Metadata     *string        `gorm:"column:metadata;type:text" json:"metadata"`
	CreatedAt    time.Time      `gorm:"column:created_at;type:timestamp without time zone;not null;default:now()" json:"created_at"`
	EventStreams []*EventStream `gorm:"foreignKey:event_id" json:"event_streams"`
}

// TableName EventStore's table name
func (*EventStore) TableName() string {
	return TableNameEventStore
}
