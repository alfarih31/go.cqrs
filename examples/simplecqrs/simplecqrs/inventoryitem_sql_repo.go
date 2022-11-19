package simplecqrs

import (
	"context"
	ycq "github.com/jetbasrawi/go.cqrs"
)

type InventoryItemSqlRepo struct {
	repo ycq.DomainRepository
}

// NewInventoryItemSqlRepo constructs a new InventoryItemSqlRepository.
func NewInventoryItemSqlRepo(repo ycq.EventRepository, eventBus ycq.EventBus) (*InventoryItemSqlRepo, error) {

	sqlDomainRepo, err := ycq.NewSqlDomainRepository(repo, eventBus)
	if err != nil {
		return nil, err
	}

	eventFactory := ycq.NewDelegateEventFactory()
	eventFactory.RegisterDelegate(string(InventoryItemCreatedEvent), func() ycq.Event {
		return NewInventoryEvent[InventoryItemCreated](InventoryItemCreated{}, string(InventoryItemCreatedEvent))
	})
	eventFactory.RegisterDelegate(string(InventoryItemRenamedEvent), func() ycq.Event {
		return NewInventoryEvent[InventoryItemRenamed](InventoryItemRenamed{}, string(InventoryItemRenamedEvent))
	})
	eventFactory.RegisterDelegate(string(InventoryItemDeactivatedEvent), func() ycq.Event {
		return NewInventoryEvent[InventoryItemDeactivated](InventoryItemDeactivated{}, string(InventoryItemDeactivatedEvent))
	})
	eventFactory.RegisterDelegate(string(ItemsRemovedFromInventoryEvent), func() ycq.Event {
		return NewInventoryEvent[ItemsRemovedFromInventory](ItemsRemovedFromInventory{}, string(ItemsRemovedFromInventoryEvent))
	})
	eventFactory.RegisterDelegate(string(ItemsCheckedIntoInventoryEvent), func() ycq.Event {
		return NewInventoryEvent[ItemsCheckedIntoInventory](ItemsCheckedIntoInventory{}, string(ItemsCheckedIntoInventoryEvent))
	})
	sqlDomainRepo.SetEventFactory(eventFactory)

	streamNameDelegate := ycq.NewDelegateStreamNamer()
	streamNameDelegate.RegisterDelegate(func(t string, id string) string {
		return t + "#" + id
	}, &InventoryItem{})
	sqlDomainRepo.SetStreamNameDelegate(streamNameDelegate)

	// An event factory creates an instance of an event given the name of an event
	// as a string.
	aggregateFactory := ycq.NewDelegateAggregateFactory()
	aggregateFactory.RegisterDelegate(&InventoryItem{},
		func(id string) ycq.AggregateRoot {
			return NewInventoryItem(id)
		})
	sqlDomainRepo.SetAggregateFactory(aggregateFactory)

	return &InventoryItemSqlRepo{
		repo: sqlDomainRepo,
	}, nil
}

// Load loads events for an aggregate.
//
// Returns an *InventoryAggregate.
func (r *InventoryItemSqlRepo) Load(ctx context.Context, aggregateTypeName, aggregateId string) (*InventoryItem, error) {
	i, err := r.repo.Load(ctx, aggregateTypeName, aggregateId)
	if err != nil {
		return nil, err
	}

	return i.(*InventoryItem), nil
}

// Save persists an aggregate.
func (r *InventoryItemSqlRepo) Save(ctx context.Context, aggregate ycq.AggregateRoot, expectedVersion *int) error {
	return r.repo.Save(ctx, aggregate, expectedVersion)
}
