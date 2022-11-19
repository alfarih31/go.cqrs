package simplecqrs

import (
	"errors"
	"github.com/jetbasrawi/go.cqrs"
)

// InventoryItem is the aggregate for an inventory item.
type InventoryItem struct {
	*ycq.AggregateBase
	activated bool
	count     int
}

var _ ycq.AggregateRoot = new(InventoryItem)

// NewInventoryItem constructs a new inventory item aggregate.
//
// Importantly it embeds a new AggregateBase.
func NewInventoryItem(id string) *InventoryItem {
	i := &InventoryItem{
		AggregateBase: ycq.NewAggregateBase(id),
	}

	return i
}

// Create raises InventoryItemCreatedEvent
func (a *InventoryItem) Create(name string) error {
	if name == "" {
		return errors.New("the name can not be empty")
	}

	a.Apply(ycq.NewEventMessage(a.AggregateID(),
		NewInventoryEvent[InventoryItemCreated](InventoryItemCreated{ID: a.AggregateID(), Name: name}, "InventoryItemCreated"),
		ycq.Int(a.CurrentVersion())), true)

	return nil
}

// ChangeName changes the name of the item.
func (a *InventoryItem) ChangeName(newName string) error {
	if newName == "" {
		return errors.New("the name can not be empty")
	}

	a.Apply(ycq.NewEventMessage(a.AggregateID(),
		NewInventoryEvent[InventoryItemRenamed](InventoryItemRenamed{ID: a.AggregateID(), NewName: newName}, "InventoryItemRenamed"),
		ycq.Int(a.CurrentVersion())), true)

	return nil
}

// Remove removes items from inventory.
func (a *InventoryItem) Remove(count int) error {
	if count <= 0 {
		return errors.New("can't remove negative count from inventory")
	}

	if a.count-count < 0 {
		return errors.New("can't remove more items from inventory than the number of items in inventory")
	}

	a.Apply(ycq.NewEventMessage(a.AggregateID(),
		NewInventoryEvent[ItemsRemovedFromInventory](ItemsRemovedFromInventory{ID: a.AggregateID(), Count: count}, "ItemsRemovedFromInventory"),
		ycq.Int(a.CurrentVersion())), true)

	return nil
}

// CheckIn adds items to inventory.
func (a *InventoryItem) CheckIn(count int) error {
	if count <= 0 {
		return errors.New("must have a count greater than 0 to add to inventory")
	}

	a.Apply(ycq.NewEventMessage(a.AggregateID(),
		NewInventoryEvent[ItemsCheckedIntoInventory](ItemsCheckedIntoInventory{ID: a.AggregateID(), Count: count}, "ItemsCheckedIntoInventory"),
		ycq.Int(a.CurrentVersion())), true)

	return nil
}

// Deactivate deactivates the inventory item.
func (a *InventoryItem) Deactivate() error {
	if !a.activated {
		return errors.New("already deactivated")
	}

	a.Apply(ycq.NewEventMessage(a.AggregateID(),
		NewInventoryEvent[InventoryItemDeactivated](InventoryItemDeactivated{ID: a.AggregateID()}, "InventoryItemDeactivated"),
		ycq.Int(a.CurrentVersion())), true)

	return nil
}

// Apply handles the logic of events on the aggregate.
func (a *InventoryItem) Apply(message ycq.EventMessage, isNew bool) {
	if isNew {
		a.TrackChange(message)
	}

	switch InventoryEventName(message.Event().Name()) {
	case InventoryItemCreatedEvent:
		a.activated = true
	case InventoryItemDeactivatedEvent:
		a.activated = false
	case ItemsRemovedFromInventoryEvent:
		d := message.Event().Data().(ItemsRemovedFromInventory)
		a.count -= d.Count
	case ItemsCheckedIntoInventoryEvent:
		d := message.Event().Data().(ItemsCheckedIntoInventory)
		a.count += d.Count
	}
}

func (a *InventoryItem) RebuildFromEvents(events []ycq.EventMessage) {
	for _, e := range events {
		a.Apply(e, false)
		a.IncrementVersion()
	}
}
