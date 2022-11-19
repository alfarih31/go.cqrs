package simplecqrs

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type InventoryEventName string

const (
	InventoryItemCreatedEvent      InventoryEventName = "InventoryItemCreated"
	InventoryItemRenamedEvent      InventoryEventName = "InventoryItemRenamed"
	InventoryItemDeactivatedEvent  InventoryEventName = "InventoryItemDeactivated"
	ItemsRemovedFromInventoryEvent InventoryEventName = "ItemsRemovedFromInventory"
	ItemsCheckedIntoInventoryEvent InventoryEventName = "ItemsCheckedIntoInventory"
)

// Events are just plain structs

// InventoryItemCreated event
type InventoryItemCreated struct {
	ID   string
	Name string
}

// InventoryItemRenamed event
type InventoryItemRenamed struct {
	ID      string
	NewName string
}

// InventoryItemDeactivated event
type InventoryItemDeactivated struct {
	ID string
}

// ItemsRemovedFromInventory event
type ItemsRemovedFromInventory struct {
	ID    string
	Count int
}

// ItemsCheckedIntoInventory event
type ItemsCheckedIntoInventory struct {
	ID    string
	Count int
}

type InventoryEvent[T any] struct {
	data T
	name string
}

func NewInventoryEvent[T any](data T, name string) *InventoryEvent[T] {
	return &InventoryEvent[T]{
		data: data,
		name: name,
	}
}

func (i *InventoryEvent[T]) Data() interface{} {
	return i.data
}

func (i *InventoryEvent[T]) Name() string {
	return i.name
}

func (i *InventoryEvent[T]) Unmarshal(rawString string) error {
	v := reflect.ValueOf(i.data)
	switch v.Kind() {
	case reflect.Ptr:
		return json.Unmarshal([]byte(rawString), i.data)
	case reflect.Struct:
		return json.Unmarshal([]byte(rawString), &i.data)
	}
	return fmt.Errorf("unknown InventoryEvent data types")
}

func (i *InventoryEvent[T]) Marshal() (string, error) {
	jb, err := json.Marshal(i.data)
	if err != nil {
		return "", err
	}

	return string(jb), nil
}
