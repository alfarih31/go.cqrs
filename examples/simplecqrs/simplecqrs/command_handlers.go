package simplecqrs

import (
	"context"
	"github.com/jetbasrawi/go.cqrs"
	"log"
)

type InventoryItemRepository interface {
	Load(context.Context, string, string) (*InventoryItem, error)
	Save(context.Context, ycq.AggregateRoot, *int) error
}

// InventoryCommandHandlers provides methods for processing commands related
// to inventory items.
type InventoryCommandHandlers struct {
	repo InventoryItemRepository
}

// NewInventoryCommandHandlers contructs a new InventoryCommandHandlers
func NewInventoryCommandHandlers(repo InventoryItemRepository) *InventoryCommandHandlers {
	return &InventoryCommandHandlers{
		repo: repo,
	}
}

// Handle processes inventory item commands.
func (h *InventoryCommandHandlers) Handle(ctx context.Context, message ycq.CommandMessage) error {

	var item *InventoryItem
	switch cmd := message.Command().(type) {
	case *CreateInventoryItem:
		item = NewInventoryItem(message.AggregateID())
		if err := item.Create(cmd.Name); err != nil {
			return &ycq.ErrCommandExecution{Command: message, Reason: err.Error()}
		}
		return h.repo.Save(ctx, item, ycq.Int(item.OriginalVersion()))

	case *DeactivateInventoryItem:

		item, _ = h.repo.Load(ctx, ycq.TypeOf(&InventoryItem{}), message.AggregateID())
		if err := item.Deactivate(); err != nil {
			return &ycq.ErrCommandExecution{Command: message, Reason: err.Error()}
		}
		return h.repo.Save(ctx, item, ycq.Int(item.OriginalVersion()))

	case *RemoveItemsFromInventory:

		item, _ = h.repo.Load(ctx, ycq.TypeOf(&InventoryItem{}), message.AggregateID())
		item.Remove(cmd.Count)
		return h.repo.Save(ctx, item, ycq.Int(item.OriginalVersion()))

	case *CheckInItemsToInventory:
		item, _ = h.repo.Load(ctx, ycq.TypeOf(&InventoryItem{}), message.AggregateID())
		item.CheckIn(cmd.Count)
		return h.repo.Save(ctx, item, ycq.Int(item.OriginalVersion()))

	case *RenameInventoryItem:

		item, err := h.repo.Load(ctx, ycq.TypeOf(&InventoryItem{}), message.AggregateID())
		if err != nil {
			return &ycq.ErrCommandExecution{Command: message, Reason: err.Error()}
		}
		if err = item.ChangeName(cmd.NewName); err != nil {
			return &ycq.ErrCommandExecution{Command: message, Reason: err.Error()}
		}
		return h.repo.Save(ctx, item, ycq.Int(item.OriginalVersion()))

	default:
		log.Fatalf("InventoryCommandHandlers has received a command that it is does not know how to handle, %#v", cmd)
	}

	return nil
}
