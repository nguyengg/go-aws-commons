package ddbfns

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Manager provides methods to execute DynamoDB requests.
//
// The zero-value is ready for use.
type Manager struct {
	// Client is the DynamoDB client to execute requests.
	//
	// If not given, the default config (`config.LoadDefaultConfig`) will be used to create the DynamoDB client.
	Client ManagerAPIClient

	// Builder is the Builder instance to use when building input parameters.
	//
	// Default to DefaultBuilder.
	Builder *Builder

	// ClientOptions are passed to each DynamoDB request.
	ClientOptions []func(*dynamodb.Options)

	init sync.Once
}

// NewManager returns a new Manager using the given client.
func NewManager(client ManagerAPIClient, optFns ...func(*Manager)) *Manager {
	m := &Manager{
		Client:        client,
		Builder:       &Builder{},
		ClientOptions: make([]func(*dynamodb.Options), 0),
	}
	for _, fn := range optFns {
		fn(m)
	}

	return m
}

// ManagerAPIClient abstracts the DynamoDB APIs that are used by Manager.
type ManagerAPIClient interface {
	DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
}

// decode is utility function to decode return values (given by itemFn) into either values or
// valuesOnConditionCheckFailure, depending on the given error.
func (m *Manager) decode(err error, values interface{}, valuesOnConditionCheckFailure interface{}, itemFn func() map[string]types.AttributeValue) error {
	if err != nil {
		if valuesOnConditionCheckFailure == nil {
			return err
		}

		var ex *types.ConditionalCheckFailedException
		if errors.As(err, &ex) {
			if item := itemFn(); len(item) != 0 {
				if err = m.Builder.Decoder.Decode(&types.AttributeValueMemberM{Value: item}, valuesOnConditionCheckFailure); err != nil {
					return fmt.Errorf("unmarshal returned values on condition check failure error: %w", err)
				}
			}
		}

		return err
	}

	if values == nil {
		return nil
	}

	if item := itemFn(); len(item) != 0 {
		if err = m.Builder.Decoder.Decode(&types.AttributeValueMemberM{Value: item}, values); err != nil {
			return fmt.Errorf("unmarshal returned values error: %w", err)
		}
	}

	return nil
}

func (m *Manager) initFn(ctx context.Context) error {
	if m.Client == nil {
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return err
		}

		m.Client = dynamodb.NewFromConfig(cfg)
	}

	if m.Builder == nil {
		m.Builder = &Builder{}
	}

	return nil
}

// DefaultManager is the zero-value Manager that uses DefaultBuilder and is used by Delete, Get, Put, and Update.
var DefaultManager = &Manager{Builder: DefaultBuilder}
